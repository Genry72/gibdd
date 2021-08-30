package utils

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

//Adddb первоначальное создание баз
func AddDB() (err error) {
	//Создаем базу если ее нет
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Создаем таблицу с пользователями
	usersTab := `
	CREATE TABLE IF NOT EXISTS users(
		chatID		INTEGER PRIMARY KEY,
		name		TEXT,
		username	INTEGER,
		create_date TEXT,
		navi_date 	TEXT
	  )
	`
	//Создаем таблицу с рег данными
	regInfoTab := `
	CREATE TABLE IF NOT EXISTS regInfo(
		id    		INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		regnum		TEXT,
		stsnum		INTEGER,
		chatID		INTEGER, --Принадлежность пользоватлею
		create_date TEXT
	  )
	`
	//Создаем таблицу с информацией об отправленных уведомлениях
	events := `
	CREATE TABLE IF NOT EXISTS events(
		id    		INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		chatID		INTEGER, --Принадлежность пользоватлею
		numberPost  TEXT, --номер постановления
		create_date TEXT,
		DateDiscount TEXT --Дата окончания действия скидки
	  )
	`
	_, err = db.Exec(usersTab)
	if err != nil {
		err = fmt.Errorf("не удальсь создать таблицу users: %v", err)
		return
	}
	_, err = db.Exec(regInfoTab)
	if err != nil {
		err = fmt.Errorf("не удальсь создать таблицу regInfo: %v", err)
		return
	}
	_, err = db.Exec(events)
	if err != nil {
		err = fmt.Errorf("не удальсь создать таблицу events: %v", err)
		return
	}
	return
}

//AddUser Добавление пользователя в БД
func AddUser(sender, username string, chatID int) (err error) {
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Проверяем существование пользователя
	est, err := checkZnachDB("users", "chatID", fmt.Sprintf("%v", chatID))
	if err != nil {
		err = fmt.Errorf("ошибка получения данных по пользователю из БД %v", err)
		return
	}
	if est { //выходим если пользоватлеь есть
		log.Println("Пользователь уже есть")
		return
	}
	log.Printf("Добавляем пользователя %v", username)
	insert := "INSERT INTO users (name, chatID, username, create_date, navi_date) VALUES ($1, $2, $3, $4, $5)"
	statement, _ := db.Prepare(insert)                                                          //Подготовка вставки
	_, err = statement.Exec(sender, chatID, username, time.Now().String(), time.Now().String()) //Вставка с параметрами
	if err != nil {
		err = fmt.Errorf("ошибка инсета в БД:%v Запрос: %v ", err, insert)
		return
	}
	infoLog.Printf("Пользователь %v добавлен в БД", username)
	return
}

//AddReg Добавление регистрационные данные в БД
func AddReg(regnum, regreg, stsnum string, chatID int) (err error) {
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Проверяем наличие регданных в БД
	est, err := chechReg(stsnum, fmt.Sprint(chatID))
	if err != nil {
		return
	}
	if est { //выходим если рег данные есть в БД
		err = fmt.Errorf("регистрационные данные были добавлены ранее")
		return
	}
	for i, proxyHost := range goodProxyList {
		//Проверяем валидность данных на сайте gibdd
		err = СheckRegNum(regnum, regreg, stsnum, proxyHost)
		if err != nil {
			if i == len(proxyHost)-1 {
				err = fmt.Errorf("не удалось проверить данные %v", err)
				return
			}
			warnLog.Println(err)
			continue
		}
		break
	}

	log.Println("Добавляем рег данные")
	insert := "INSERT INTO reginfo (regnum, stsnum, chatID, create_date) VALUES ($1, $2, $3, $4)"
	statement, _ := db.Prepare(insert)                                                             //Подготовка вставки
	_, err = statement.Exec(regnum+regreg, stsnum, fmt.Sprintf("%v", chatID), time.Now().String()) //Вставка с параметрами
	if err != nil {
		err = fmt.Errorf("ошибка инсета в БД:%v Запрос: %v ", err, insert)
		return
	}
	infoLog.Println("Рег данные успешно добавлены")
	return
}

//checkZnachDB Проверка существование записи в БД
func checkZnachDB(tableName, znachName, znach string) (est bool, err error) {
	est = false
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	qwery := fmt.Sprintf("select count (*) from %v where %v = %v", tableName, znachName, znach)
	row := db.QueryRow(qwery)
	var result string
	err = row.Scan(&result)
	if err != nil {
		err = fmt.Errorf("ошибка выполнения единичного запроса в БД %v: %v", qwery, err)
		return
	}
	if result != "0" {
		est = true
	}
	return
}

//chechUser Проверка существование рег данных в БД
func chechReg(stsNum, chatID string) (est bool, err error) {
	est = false
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	qwery := fmt.Sprintf("select count (*) from reginfo where stsNum = %v and chatID = %v", stsNum, chatID)
	row := db.QueryRow(qwery)
	var result string
	err = row.Scan(&result)
	if err != nil {
		err = fmt.Errorf("ошибка выполнения единичного запроса в БД %v: %v", qwery, err)
		return
	}
	if result != "0" {
		est = true
	}
	return
}

//Getreg Возвращает мапу с данными для получения штрафов check - проверка по запросу или штатная
func Getreg() (mapa map[int][]string, err error) {

	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		log.Fatal(err)
		return
	}

	defer db.Close()
	rows, err := db.Query("SELECT id, chatID, regnum, stsnum FROM reginfo")
	if err != nil {
		log.Fatal(err)
	}

	var id int
	var chatID string
	var regnum string
	var stsnum string
	checkMap := make(map[int][]string) //Мапа для проверки штрафов по всем рег данным
	for rows.Next() {
		err = rows.Scan(&id, &chatID, &regnum, &stsnum)
		if err != nil {
			log.Fatal(err)
		}
		checkMap[id] = []string{chatID, regnum, stsnum}
	}
	mapa = checkMap
	return
}

//AddEvent Добавляем дату отправки уведомления
func AddEvent(chatID int, numberPost, DateDiscount string) (err error) {
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Перед вставкой, проверяем наличие записи
	est, err := СheckEvent(chatID, numberPost)
	if err != nil {
		return
	}
	if est {
		log.Println("Уведомление уже было")
		return
	}
	log.Println("Добавляем дату отправки уведомления")
	insert := "INSERT INTO events (chatID, numberPost, create_date, DateDiscount) VALUES (?, ?, ?, ?)"
	statement, _ := db.Prepare(insert)                                             //Подготовка вставки
	_, err = statement.Exec(chatID, numberPost, time.Now().String(), DateDiscount) //Вставка с параметрами
	if err != nil {
		err = fmt.Errorf("ошибка инсета в БД:%v Запрос: %v ", err, insert)
		return
	}
	infoLog.Println("Добавлена дата отправки уведомления")
	return
}

//checkEvent проверяет наличие уведомления в БД
func СheckEvent(chatID int, numberPost string) (est bool, err error) {
	est = false
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	qwery := fmt.Sprintf("select count (*) from events where chatID = %v and numberPost = \"%v\"", chatID, numberPost)
	infoLog.Println(qwery)
	row := db.QueryRow(qwery)
	var result string
	err = row.Scan(&result)
	if err != nil {
		err = fmt.Errorf("ошибка выполнения единичного запроса в БД %v: %v", qwery, err)
		return
	}
	if result != "0" {
		est = true
	}
	return
}

//GetDiscount Возвращает мапу для оправки уведомлений, о том что заканчивается период оплаты со скидкой
func GetDiscount() (mapa map[string][]string, err error) {
	db, err := sql.Open("sqlite3", "./yadisk/sync/gibddBot/gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}

	defer db.Close()
	zapros := fmt.Sprintf("SELECT chatID, numberPost, DateDiscount FROM events WHERE DateDiscount BETWEEN '%v' AND '%v';", time.Now().String(), time.Now().Add(72*time.Hour).String())
	rows, err := db.Query(zapros)
	if err != nil {
		return
	}

	var numberPost string
	var chatID string
	var DateDiscount string
	checkMap := make(map[string][]string) //Мапа для проверки штрафов по всем рег данным
	for rows.Next() {
		err = rows.Scan(&chatID, &numberPost, &DateDiscount)
		if err != nil {
			return
		}
		checkMap[numberPost] = []string{DateDiscount, chatID}
	}
	mapa = checkMap
	return
}
