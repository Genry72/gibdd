package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gibdd/telegram"
	"gibdd/utils"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//тоду
//Добавить колонку с временем, до какого числа можно оплатить со скидкой и реализовать функцию по уведомлению заранее.
func main() {
	myID, _ := strconv.Atoi(os.Getenv("myIDtelega"))
	if myID == 0 {
		err := fmt.Errorf("не задан myID")
		log.Fatal(err)
	}
	// nomer := "Х752ТТ"
	// region := "152"
	// sts := "9933143213"
	//Создаем необходимые БД
	err := utils.AddDB()
	if err != nil {
		log.Fatal(err)
	}
	msg := "Я стартанул"
	log.Println(msg)
	telegram.SendMessage(msg, myID)
	go func() {
		for {
			//Получаем офсет
			offset, err := telegram.GetOfset()
			if err != nil {
				log.Printf("Ошибка получения офсета %v\n", err)
				time.Sleep(5 * time.Minute)
				continue
			}
			if offset == 0 {
				continue
			}
			for {
				message, sender, chatID, newOffset, username, err := telegram.Getupdate(offset)
				if err != nil {
					log.Printf("Ошибка получения сообщения %v\n", err)
					time.Sleep(5 * time.Minute)
					continue
				}
				if message == "" {
					offset = newOffset
					continue
				}
				log.Printf("Получено сообщение от пользователя %v chatID:%v http://t.me/%v с текстом: %v", sender, chatID, username, message)
				telegram.SendMessage(fmt.Sprintf("Debug: Получено сообщение от пользователя %v chatID:%v http://t.me/%v с текстом: %v", sender, chatID, username, message), myID) //Все сообщения боту для дебага мне

				command := strings.Split(message, " ") //бьем пробелами
				switch strings.ToUpper(command[0]) {   //Берем первое значение

				case "РЕГ":
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						log.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
					reg := strings.Split(strings.ToUpper(command[1]), ":") //Получаем рег данные
					fullRegnum := reg[0]                                   //Полный номер, включая регион
					regnum := string([]rune(fullRegnum)[:6])               //Первые 6 символов (номер)
					regreg := string([]rune(fullRegnum)[6:])               //Обрезаем первые 6 символов (регион)
					stsnum := reg[1]

					//Проверяем валидность на сайте gibdd
					err = getShtrafs(regnum, regreg, stsnum, chatID, false)
					if err != nil {
						log.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
						if err.Error() == "рег данные уже есть" {
							telegram.SendMessage(fmt.Sprintf("%s", err), chatID)
						} else {
							telegram.SendMessage("Не найдено ТС с таким сочетанием СТС и ГРЗ", chatID)
						}
						break
					}
					//Добавляем рег данные в БД
					// err = utils.AddReg(fullRegnum, stsnum, chatID)
					// if err != nil {
					// 	log.Println(err)
					// 	telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					// 	if err.Error() == "рег данные уже есть" {
					// 		telegram.SendMessage(fmt.Sprintf("%s", err), chatID)
					// 	} else {
					// 		telegram.SendMessage("Не удалось добавить регистрационные данные", chatID)
					// 	}
					// 	break
					// }
					telegram.SendMessage("Debug: Рег данные успешно добавлены", myID)
					telegram.SendMessage("Данные успешно добавлены", chatID)
				case "/START", "/HELP":
					telegram.SendMessage(`
Бот находится на этапе разрабоки!
Контактные данные http://t.me/valentinsemenov
Для добавления регистрационных данных отправьте:
рег A999AA555:1111111111
Где:
А999АА          госномер в русской расскладке
555                 регион
1111111111  Свидетельство о регистрации (СТС)
`, chatID)
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						log.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
				case "/CHECK":
					printShtraf(myID, true, chatID)
				default:
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						log.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
					telegram.SendMessage("Debug: Не корректная команда, наберите /help для справки", myID)
					telegram.SendMessage("Не корректная команда, наберите /help для справки", chatID)
				}
				offset = newOffset
			}
		}
	}()
	// time.Sleep(60 * time.Minute)
	//Проверяем штрафы
	for {
		printShtraf(myID, false, 0)
		time.Sleep(30 * time.Second)
	}

	// time.Sleep(60 * time.Minute)
}

//printShtraf Печатаем штрафы в телегу
func printShtraf(myID int, check bool, currentChatID int) {
	//Получаем мапу с данными для проверки штрафов
	mapa, err := utils.Getreg()
	if err != nil {
		msg := fmt.Sprintf("Debug: ошибка получения мапы: %v", err)
		telegram.SendMessage(msg, myID)
		log.Println(msg)
		return
	}
	//Вызываем проверку
	for _, regs := range mapa {
		chatID := regs[0]
		id, _ := strconv.Atoi(chatID)
		if check {
			if chatID != fmt.Sprint(currentChatID) {
				continue
			}
		}
		fullRegnum := regs[1]                    //Полный номер, включая регион
		nomer := string([]rune(fullRegnum)[:6])  //Первые 6 символов (номер)
		region := string([]rune(fullRegnum)[6:]) //Обрезаем первые 6 символов (регион)
		sts := regs[2]

		err = getShtrafs(nomer, region, sts, id, check)
		if err != nil {
			log.Println(err)
			telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
		}
	}

}

//getShtrafs Функция отправляет штрафы по конкретному пользователю + ПТС
func getShtrafs(nomer, region, sts string, chatID int, check bool) (err error) {
	log.Println("Получаем штрафы")
	// var shtrafs []string
	// myID, _ := strconv.Atoi(os.Getenv("myIDtelega"))
	url := "https://check.gibdd.ru/proxy/check/fines"
	method := "POST"
	payload := strings.NewReader("regnum=" + nomer + "&regreg=" + region + "&stsnum=" + sts)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("выполнение запроса %v %v завершиблось ошибкой %v: %v", url, payload, res.Status, string(body))
		return
	}
	m := shtrafStrukt{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение списка штрафов: %v Боди: %v", err, string(body))
		return
	}
	var post string
	var divid int
	if m.Code != 200 {
		err = fmt.Errorf("выполнение запроса %v %v завершиблось ошибкой %v: %v", url, payload, m.Message, string(body))
		return
	}
	//Добавляем рег данные, в случае если запрос выше отработал
	err = utils.AddReg(nomer+region, sts, chatID, check)
	if err != nil {
		if err.Error() != "рег данные уже есть" { //Выходим если ошибка отличная от этой
			return
		}

	}
	cafapPicsToken := m.CafapPicsToken
	for _, shtraf := range m.Data {
		dateNarush := shtraf.DateDecis
		// fmt.Printf("Штраф на сумму %vр, со скидкой можно опатить до %v\n", shtraf.Summa, shtraf.DateDiscount)
		// fmt.Printf("Дата нарушения %v\n", dateNarush)
		shtrafString := fmt.Sprintf("❗Штраф на сумму %vр\n", shtraf.Summa)
		shtrafString = shtrafString + fmt.Sprintf("Оплата со скидкой до %v\n", shtraf.DateDiscount)
		shtrafString = shtrafString + fmt.Sprintf("Дата нарушения %v\n", dateNarush)
		shtrafString = shtrafString + fmt.Sprintf("Адрес: %v\n", m.Divisions[shtraf.Division]["fulladdr"])

		post = shtraf.NumPost
		divid = shtraf.Division
		//Проверяем, были ли ранее уведомления, в случае, если это проверки по циклу
		if !check {
			est, err := utils.СheckEvent(chatID, post)
			if err != nil {
				return err
			}
			if est {
				log.Println("уведомление уже было")
				continue
			}
		}
		countPhoto, err := linkImage(post, nomer+region, fmt.Sprintf("%v", divid), cafapPicsToken, shtraf.NumPost+".jpeg")
		if err != nil {
			err = fmt.Errorf("ошибка получения картинки со штрафом: %v", err)
			log.Println(err)
			shtrafString = shtrafString + "Фото штрафа не загружено"
			// shtrafs = append(shtrafs, shtrafString)
		}
		// shtrafs = append(shtrafs, shtrafString)
		var errSend = false //Если хотябы одна картинка не отправилась, то считаем что уведомление не ушло
		for i := 0; i < countPhoto; i++ {
			var msg string
			if i == countPhoto-1 {
				msg = shtrafString
			}
			// telegram.SendPhoto(fmt.Sprint(i)+shtraf.NumPost+".jpeg", "Debug: "+msg, myID)
			err = telegram.SendPhoto(fmt.Sprint(i)+shtraf.NumPost+".jpeg", msg, chatID)
			if err != nil {
				log.Printf("ошибка отправки фото: %v", err)
				errSend = true
			}
			os.Remove("./" + fmt.Sprint(i) + shtraf.NumPost + ".jpeg")
		}
		if !errSend { //Если при отправки не было ошибок, то добавляем запись
			err = utils.AddEvent(chatID, post)
			if err != nil {
				return err
			}
		}

	}
	return
}

//linkImage Получаем ссылку на картинку
func linkImage(post, regnum, divid, cafapPicsToken, filename string) (countPhoto int, err error) {
	url := "https://check.gibdd.ru/proxy/check/fines/pics"
	method := "POST"
	payload := strings.NewReader("post=" + post + "&regnum=" + regnum + "&divid=" + divid + "&cafapPicsToken=" + cafapPicsToken)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("выполнение запроса %v завершиблось ошибкой %v: %v", url, res.Status, string(body))
		return
	}
	m := linkImageStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение картинки штрафа: %v Боди: %v", err, string(body))
		return
	}
	if len(m.Photos) == 0 {
		err = fmt.Errorf("нет урла штрафа: %v", string(body))
		return
	}
	countPhoto = len(m.Photos)
	for i := 0; i < countPhoto; i++ {
		link := fmt.Sprintf("%v", m.Photos[i].Base64Value)
		err = base64toJpg("./"+fmt.Sprint(i)+filename, link)
	}
	return
}

func base64toJpg(filepath, data string) (err error) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		return err
	}
	//Encode from image format to writer
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	err = jpeg.Encode(f, m, &jpeg.Options{Quality: 75})
	if err != nil {
		return err
	}
	return err
}

type shtrafStrukt struct {
	DurationReg int    `json:"durationReg"`
	Request     string `json:"request"`
	Code        int    `json:"code"`
	Data        []struct {
		Discount       string `json:"Discount"`
		EnableDiscount bool   `json:"enableDiscount"`
		DateDecis      string `json:"DateDecis"`
		KoAPcode       string `json:"KoAPcode"`
		DateDiscount   string `json:"DateDiscount"`
		VehicleModel   string `json:"VehicleModel"`
		KoAPtext       string `json:"KoAPtext"`
		NumPost        string `json:"NumPost"`
		Kbk            string `json:"kbk"`
		Summa          int    `json:"Summa"`
		Division       int    `json:"Division"`
		EnablePics     bool   `json:"enablePics"`
		ID             string `json:"id"`
		SupplierBillID string `json:"SupplierBillID"`
		DatePost       string `json:"DatePost"`
	} `json:"data"`
	EndDate        string                         `json:"endDate"`
	CafapPicsToken string                         `json:"cafapPicsToken"`
	Message        string                         `json:"message"`
	Divisions      map[int]map[string]interface{} `json:"divisions"`
	RequestTime    string                         `json:"requestTime"`
	Duration       int                            `json:"duration"`
	Hostname       string                         `json:"hostname"`
	MessageReg     string                         `json:"messageReg"`
	StartDate      string                         `json:"startDate"`
}

type linkImageStruct struct {
	RequestTime string `json:"requestTime"`
	Hostname    string `json:"hostname"`
	Code        string `json:"code"`
	ReqToken    string `json:"reqToken"`
	Comment     string `json:"comment"`
	Photos      []struct {
		Base64Value string `json:"base64Value"`
		Type        int    `json:"type"`
	} `json:"photos"`
	Version string `json:"version"`
}
