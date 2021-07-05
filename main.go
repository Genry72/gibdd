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
				log.Printf("Получено сообщение от пользователя %v chatID:%v username:%v с текстом: %v", sender, chatID, username, message)
				telegram.SendMessage(fmt.Sprintf("Debug: Получено сообщение от пользователя %v chatID:%v username:%v с текстом: %v", sender, chatID, username, message), myID) //Все сообщения боту для дебага мне

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
					err = checkShtraf(regnum, regreg, stsnum, fmt.Sprint(chatID), fullRegnum, myID)
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
		func() {
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
				fullRegnum := regs[1]                    //Полный номер, включая регион
				nomer := string([]rune(fullRegnum)[:6])  //Первые 6 символов (номер)
				region := string([]rune(fullRegnum)[6:]) //Обрезаем первые 6 символов (регион)
				sts := regs[2]

				err = checkShtraf(nomer, region, sts, chatID, fullRegnum, myID)
				if err != nil {
					log.Println(err)
					telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
				}
			}

		}()
		time.Sleep(5 * time.Minute)
	}

	// time.Sleep(60 * time.Minute)
}

//checkShtraf Функция получения мапы со штрафами
func getShtrafs(nomer, region, sts string) (shtrafs []string, err error) {
	log.Println("Получаем мапу со штрафами")
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
		err = fmt.Errorf("выполнение запроса %v завершиблось ошибкой %v: %v", url, res.Status, string(body))
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
		err = fmt.Errorf("выполнение запроса %v завершиблось ошибкой %v: %v", url, m.Message, string(body))
		return
	}
	cafapPicsToken := m.CafapPicsToken
	for _, shtraf := range m.Data {
		dateNarush := shtraf.DateDecis
		// fmt.Printf("Штраф на сумму %vр, со скидкой можно опатить до %v\n", shtraf.Summa, shtraf.DateDiscount)
		// fmt.Printf("Дата нарушения %v\n", dateNarush)
		shtrafString := fmt.Sprintf("Штраф на сумму %vр, со скидкой можно опатить до %v\n", shtraf.Summa, shtraf.DateDiscount)
		shtrafString = shtrafString + fmt.Sprintf("Дата нарушения %v\n", dateNarush)
		shtrafs = append(shtrafs, shtrafString)
		post = shtraf.NumPost
		divid = shtraf.Division
		err = linkImage(post, nomer+region, fmt.Sprintf("%v", divid), cafapPicsToken, shtraf.NumPost+".jpeg")
		if err != nil {
			err = fmt.Errorf("ошибка получения картинки со штрафом: %v", err)
			log.Println(err)
			return shtrafs, nil
		}
	}
	// fmt.Println(string(body))
	log.Println("Мапа со штрафами получена")
	return
}

//checkShtraf выводим штрафы
func checkShtraf(nomer, region, sts, chatID, fullRegnum string, myID int) (err error) {
	log.Println("Проверяем штрафы")
	id, _ := strconv.Atoi(chatID)
	err = utils.AddEvent(fullRegnum, sts, id)
	if err != nil {
		return
	}
	shtrafs, err := getShtrafs(nomer, region, sts)
	if err != nil {
		err = fmt.Errorf("ошибка при получении штрафов: %v", err)
		return
	}
	for _, shtrafString := range shtrafs {
		telegram.SendMessage(fmt.Sprintf("Debug: %v", shtrafString), myID)
		id, _ := strconv.Atoi(chatID)
		telegram.SendMessage(shtrafString, id)
	}
	log.Println("Штрафы получены")
	return
}

//linkImage Получаем ссылку на картинку
func linkImage(post, regnum, divid, cafapPicsToken, filename string) (err error) {
	url := "https://check.gibdd.ru/proxy/check/fines/pics"
	method := "POST"
	payload := strings.NewReader("post=" + post + "&regnum=" + regnum + "&divid=" + divid + "&cafapPicsToken=" + cafapPicsToken)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("выполнение запроса %v завершиблось ошибкой %v: %v", url, res.Status, string(body))
		return err
	}
	m := linkImageStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение картинки штрафа: %v Боди: %v", err, string(body))
		return err
	}
	if len(m.Photos) == 0 {
		err = fmt.Errorf("нет урла штрафа: %v", string(body))
		return err
	}
	link := fmt.Sprintf("%v", m.Photos[0].Base64Value)
	err = base64toJpg("./"+filename, link)
	return err
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
	EndDate        string `json:"endDate"`
	CafapPicsToken string `json:"cafapPicsToken"`
	Message        string `json:"message"`
	Divisions      struct {
		Num1109800 struct {
			Regkod   string        `json:"regkod"`
			Activity []interface{} `json:"activity"`
			Name     string        `json:"name"`
			SiteID   int           `json:"siteId"`
			Active   bool          `json:"active"`
			Fulladdr string        `json:"fulladdr"`
			DivID    int           `json:"divId"`
			Coords   string        `json:"coords"`
		} `json:"1109800"`
	} `json:"divisions"`
	RequestTime string `json:"requestTime"`
	Duration    int    `json:"duration"`
	Hostname    string `json:"hostname"`
	MessageReg  string `json:"messageReg"`
	StartDate   string `json:"startDate"`
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
