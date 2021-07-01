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
	db, err := sql.Open("sqlite3", "./gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Создаем таблицу со всеми данными
	usersTab := `
	CREATE TABLE IF NOT EXISTS users(
		id    		INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		chatID		INTEGER,
		name		TEXT,
		username	TEXT,
		regnum		TEXT,
		stsnum		INTEGER UNIQUE,
		create_date TEXT,
		navi_date 	TEXT
	  )
	`
	_, err = db.Exec(usersTab)
	if err != nil {
		err = fmt.Errorf("не удальсь создать таблицу users")
		return
	}
	return
}

//AddUser Добавление пользователя в БД
func AddUser(sender, username string, chatID int) (err error) {
	db, err := sql.Open("sqlite3", "./gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Проверяем существование пользователя
	est, err := checkZnachDB("users", "chatID", fmt.Sprintf("%v", chatID))
	if est { //выходим если пользоватлеь есть
		fmt.Println("Пользоватлель уже есть")
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
	log.Printf("Пользователь %v добавлен в БД", username)
	return
}

//AddReg Добавление регистрационные данные в БД
func AddReg(regnum, stsnum string, chatID int) (err error) {
	db, err := sql.Open("sqlite3", "./gibdd.db")
	if err != nil {
		err = fmt.Errorf("ошибка создания БД:%v", err)
		return
	}
	defer db.Close()
	//Проверяем существование регданных по СТС
	est, err := checkZnachDB("reginfo", "stsnum", stsnum)
	if est { //выходим если пользоватлеь есть

		fmt.Println("Рег данные уже есть")
		return
	}
	log.Println("Добавляем рег данные")
	insert := "INSERT INTO reginfo (regnum, stsnum, chatID, create_date, navi_date) VALUES ($1, $2, $3, $4, $5)"
	statement, _ := db.Prepare(insert)                                                                           //Подготовка вставки
	_, err = statement.Exec(regnum, stsnum, fmt.Sprintf("%v", chatID), time.Now().String(), time.Now().String()) //Вставка с параметрами
	if err != nil {
		err = fmt.Errorf("ошибка инсета в БД:%v Запрос: %v ", err, insert)
		return
	}
	log.Println("Рег данные успешно добавлены")
	return
}

//checkZnachDB Проверка существование записи в БД
func checkZnachDB(tableName, znachName, znach string) (est bool, err error) {
	est = false
	db, err := sql.Open("sqlite3", "./gibdd.db")
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
