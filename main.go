package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"gibdd/telegram"
	"gibdd/utils"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var reset = "\033[0m"
var infoLog = log.New(os.Stdout, fmt.Sprint(string(colorGreen), "INFO\t"+reset), log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, fmt.Sprint(string(colorRed), "ERROR\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
var warnLog = log.New(os.Stdout, fmt.Sprint(string(colorYellow), "WARN\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)

//тоду
//команду check нужно выдавать на основе данных из бд, указывая дату последней проверки
//долгие ответы обернуть "проверяем, ожидайте"
func main() {
	myID, _ := strconv.Atoi(os.Getenv("myIDtelega"))
	if myID == 0 {
		err := fmt.Errorf("не задан myID")
		errorLog.Fatal(err)
	}
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err := fmt.Errorf("не задан токен")
		errorLog.Fatal(err)
	}
	var c string
	var test string
	flag.StringVar(&c, "c", "", "команда")
	flag.StringVar(&test, "test", "true", "Куда катим")
	flag.Parse()
	if c == "install" { //Первичная установка
		utils.Docker(c, test)
		return
	}
	if c == "update" { //Обновляем
		utils.Docker(c, test)
		return
	}
	if c == "yandex" { //Первичная установка
		utils.Docker(c, test)
		return
	}
	//Проверяем готовность диска
	err := utils.CheckYadisk()
	if err != nil {
		telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
		errorLog.Fatal(err)
	}
	//Создаем необходимые БД
	err = utils.AddDB()
	if err != nil {
		errorLog.Fatal(err)
	}
	msg := "Я стартанул"
	infoLog.Println(msg)
	telegram.SendMessage(msg, myID)
	go func() {
		for {
			//Получаем офсет
			offset, err := telegram.GetOfset()
			if err != nil {
				errorLog.Printf("Ошибка получения офсета %v ждем 5 минут\n", err)
				time.Sleep(5 * time.Minute)
				continue
			}
			// infoLog.Printf("Получили офсет %v", offset)
			if offset == 0 {
				continue
			}
			for { //Ловим все сообщения из телеги
				message, sender, chatID, newOffset, username, err := telegram.Getupdate(offset)
				if err != nil {
					errorLog.Printf("Ошибка получения сообщения %v\n", err)
					time.Sleep(5 * time.Minute)
					continue
				}
				if message == "" {
					offset = newOffset
					continue
				}
				infoLog.Printf("Получено сообщение от пользователя %v chatID:%v http://t.me/%v с текстом: %v", sender, chatID, username, message)
				telegram.SendMessage(fmt.Sprintf("Debug: Получено сообщение от пользователя %v chatID:%v http://t.me/%v с текстом: %v", sender, chatID, username, message), myID) //Все сообщения боту для дебага мне

				command := strings.Split(message, " ") //бьем пробелами
				switch strings.ToUpper(command[0]) {   //Берем первое значение

				case "РЕГ": //Добавление регистрационных данных
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
					//Добавляем рег данные
					reg := strings.Split(strings.ToUpper(command[1]), ":") //Получаем рег данные (бьем разделителем)
					fullRegnum := reg[0]                                   //Полный номер, включая регион
					regnum := string([]rune(fullRegnum)[:6])               //Первые 6 символов (номер)
					regreg := string([]rune(fullRegnum)[6:])               //Обрезаем первые 6 символов (регион)
					stsnum := reg[1]
					err = utils.AddReg(regnum, regreg, stsnum, chatID)
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
						telegram.SendMessage(fmt.Sprintf("Упс... %v", err), chatID)
						break
					}
					infoLog.Println("Регистрационные данные успешно добавлены")
					telegram.SendMessage("Debug: Регистрационные данные успешно добавлены", myID)
					telegram.SendMessage("Регистрационные данные успешно добавлены", chatID)
					//После добавления сразу делаем проверку штрафов
					proxylist, err := utils.Proxy()
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
					}
					var countShtaf int
					for i, proxyHost := range proxylist {
						countShtaf, err = sendShtafs(regnum, regreg, stsnum, chatID, true, proxyHost)
						if err != nil {
							if i == len(proxylist)-1 {
								errorLog.Println(err)
								telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
								telegram.SendMessage(fmt.Sprintf("Упс... %v", err), chatID)
								break
							}
							warnLog.Println(err)
							// telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
							// telegram.SendMessage(fmt.Sprintf("Упс... %v", err), chatID)
							continue
						}
						warnLog.Println("все ок, выходим")
						break
					}
					if err != nil {
						break
					}

					if countShtaf == 0 {
						telegram.SendMessage(fmt.Sprintf("Debug: ✅ По регистрационному номеру %v штрафов не найдено", fullRegnum), myID)
						telegram.SendMessage(fmt.Sprintf("✅ По регистрационному номеру %v штрафов не найдено", fullRegnum), chatID)
						break
					}
					telegram.SendMessage(fmt.Sprintf("Debug: ❗️❗️Колличество штрафов по номеру %v: %v", fullRegnum, countShtaf), myID)
					telegram.SendMessage(fmt.Sprintf("❗️❗️Колличество штрафов по номеру %v: %v", fullRegnum, countShtaf), chatID)
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
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
				case "/CHECK":
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
					err = printShtraf(myID, true, chatID)
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: Сорян, на текущий момент есть проблемы с доступностью сайта gibdd:%v", err), myID)
						telegram.SendMessage(fmt.Sprintln("Сорян, на текущий момент есть проблемы с доступностью сайта gibdd"), chatID)
					}
				default:
					//Сразу добавляем пользователя в базу
					err := utils.AddUser(sender, username, chatID)
					if err != nil {
						errorLog.Println(err)
						telegram.SendMessage(fmt.Sprintf("Debug: %s", err), myID)
					}
					telegram.SendMessage("Debug: Не корректная команда, наберите /help для справки", myID)
					telegram.SendMessage("Не корректная команда, наберите /help для справки", chatID)
				}
				offset = newOffset
			}
		}
	}()
	go func() { //Проверяем дискаунты
		for {
			if time.Now().Hour() != 17 { //Уведомления отправляем в 17 часу
				log.Println("Пропускаем проверку дискаунтов")
				time.Sleep(15 * time.Minute)
				continue
			}
			err = sendDiscounts(myID)
			if err != nil {
				errorLog.Printf("Ошибка проверки дискаунтов: %v", err)
				telegram.SendMessage(fmt.Sprintf("Debug: Ошибка проверки дискаунтов: %v", err), myID)
				continue
			}
			time.Sleep(1 * time.Hour) //Спим час, в случае отправки
		}

	}()
	go func() { //Раз в 5 минут обновляем список прокси серверов
		for {
			err = utils.UpdateProxyList()
			if err != nil {
				errorLog.Println(err)
				telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
				continue
			}
			time.Sleep(5 * time.Minute)
		}
	}()
	//В бесконечном цикле проверяем штрафы
	for {
		err = printShtraf(myID, false, 0)
		if err != nil {
			errorLog.Println(err)
			time.Sleep(60 * time.Minute)
		}
		time.Sleep(5 * time.Minute)
	}

	// time.Sleep(60 * time.Minute)
}

//printShtraf Печатаем штрафы в телегу
func printShtraf(myID int, check bool, currentChatID int) (err error) {
	var wg sync.WaitGroup
	//Получаем мапу с данными для проверки штрафов
	mapa, err := utils.Getreg()
	if err != nil {
		msg := fmt.Sprintf("Debug: ошибка получения мапы: %v", err)
		telegram.SendMessage(msg, myID)
		errorLog.Println(msg)
		return
	}
	//Вызываем проверку
	for _, regs := range mapa {
		chatID := regs[0]
		id, _ := strconv.Atoi(chatID)
		if check { //При запуске принудительной проверки, проверяем только по своему chatID
			if chatID != fmt.Sprint(currentChatID) {
				continue
			}
		}
		wg.Add(1)
		go func(chatID string, regs []string) error {
			fullRegnum := regs[1]                    //Полный номер, включая регион
			nomer := string([]rune(fullRegnum)[:6])  //Первые 6 символов (номер)
			region := string([]rune(fullRegnum)[6:]) //Обрезаем первые 6 символов (регион)
			sts := regs[2]
			proxylist, err := utils.Proxy()
			if err != nil {
				errorLog.Println(err)
				telegram.SendMessage(fmt.Sprintf("Debug: %v", err), myID)
			}
			var countShtaf int
			for i, proxyHost := range proxylist {
				countShtaf, err = sendShtafs(nomer, region, sts, id, check, proxyHost)
				if err != nil {
					if i == len(proxylist)-1 {
						errorLog.Printf("%v проверка %v из %v", err, i+1, len(proxylist))
						telegram.SendMessage(fmt.Sprintf("Debug: %s следующая проверка через час", err), myID)
						wg.Done()
						return err
					}
					warnLog.Printf("%v проверка %v из %v", err, i+1, len(proxylist))
					continue
				}
				break
			}

			if check {
				if countShtaf == 0 {
					telegram.SendMessage(fmt.Sprintf("Debug: ✅ По регистрационному номеру %v штрафов не найдено", fullRegnum), myID)
					telegram.SendMessage(fmt.Sprintf("✅ По регистрационному номеру %v штрафов не найдено", fullRegnum), currentChatID)
					wg.Done()
					return err
				}
				telegram.SendMessage(fmt.Sprintf("Debug: ❗️❗️ Колличество штрафов по номеру %v: %v", fullRegnum, countShtaf), myID)
				telegram.SendMessage(fmt.Sprintf("❗️❗️ Колличество штрафов по номеру %v: %v", fullRegnum, countShtaf), currentChatID)
			}
			wg.Done()
			return err
		}(chatID, regs)
	}
	wg.Wait()
	infoLog.Println("Проверка штрафов по всем рег. данным прошла")
	return
}

//sendShtafs Функция отправляет штрафы по конкретному пользователю + ПТС
func sendShtafs(nomer, region, sts string, chatID int, check bool, proxyHost string) (countShtaf int, err error) {
	if check {
		infoLog.Printf("Запрошенная проверка штрафов для %v%v:%v", nomer, region, sts)
	} else {
		infoLog.Printf("Периодическая проверка штрафов для %v%v:%v", nomer, region, sts)
	}
	//Задаем прокси
	proxyStr := "http://" + proxyHost
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return
	}
	infoLog.Printf("Используем проксю2 %v", proxyHost)
	// basicAuth := "Basic " + logpassAdLong
	// hdr := http.Header{}
	// hdr.Add("Proxy-Authorization", basicAuth)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// ProxyConnectHeader: hdr,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
	url := "https://check.gibdd.ru/proxy/check/fines"
	method := "POST"
	payload := strings.NewReader("regnum=" + nomer + "&regreg=" + region + "&stsnum=" + sts)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		err = fmt.Errorf("ошибка sendShtafs: %v", err)
		return
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("ошибка проверки штрафа для %v%v:%v %v", nomer, region, sts, err)
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
	m := utils.ShtrafStrukt{}
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
	countShtaf = len(m.Data)
	if check {
		if countShtaf == 0 {
			infoLog.Printf("Успешная проверка штрафов для %v%v:%v", nomer, region, sts)
			return
		}
	}
	cafapPicsToken := m.CafapPicsToken
	for _, shtraf := range m.Data {
		dateNarush := shtraf.DateDecis
		// fmt.Printf("Штраф на сумму %vр, со скидкой можно опатить до %v\n", shtraf.Summa, shtraf.DateDiscount)
		// fmt.Printf("Дата нарушения %v\n", dateNarush)
		shtrafString := fmt.Sprintf("❗Штраф на сумму %vр\n", shtraf.Summa)
		shtrafString = shtrafString + fmt.Sprintf("Авто %v %v\n", shtraf.VehicleModel, nomer+region)
		shtrafString = shtrafString + fmt.Sprintf("Постановление № %v\n", shtraf.NumPost)
		shtrafString = shtrafString + fmt.Sprintf("Оплата со скидкой до %v\n", shtraf.DateDiscount)
		shtrafString = shtrafString + fmt.Sprintf("Дата нарушения %v\n", dateNarush)
		shtrafString = shtrafString + fmt.Sprintf("Адрес: %v\n", m.Divisions[shtraf.Division]["fulladdr"])

		post = shtraf.NumPost
		divid = shtraf.Division
		//Проверяем, были ли ранее уведомления, в случае, если это проверки по циклу
		if !check {
			est, err := utils.СheckEvent(chatID, post)
			if err != nil {
				return countShtaf, err
			}
			if est {
				log.Println("уведомление уже было")
				continue
			}
		}
		countPhoto, err := linkImage(post, nomer+region, fmt.Sprintf("%v", divid), cafapPicsToken, shtraf.NumPost+".jpeg")
		if err != nil {
			err = fmt.Errorf("ошибка получения картинки со штрафом: %v", err)
			errorLog.Println(err)
			shtrafString = shtrafString + "Фото штрафа не загружено"
			// shtrafs = append(shtrafs, shtrafString)
		}
		var errSend = false  //Если хотябы одна картинка не отправилась, то считаем что уведомление не ушло
		if countPhoto == 0 { //Если фото нет, либо не прогрузились, отправляем как есть
			err = telegram.SendMessage(shtrafString, chatID)
			if err != nil {
				errorLog.Printf("ошибка отправки фото: %v", err)
				errSend = true //не будем добавлять инфу об отправке в БД
			}
		}
		for i := 0; i < countPhoto; i++ { //Отправояем по одной фотке, к последней прикрепляем текст
			var msg string
			if i == countPhoto-1 {
				msg = shtrafString
			}
			// telegram.SendPhoto(fmt.Sprint(i)+shtraf.NumPost+".jpeg", "Debug: "+msg, myID)
			err = telegram.SendPhoto(fmt.Sprint(i)+shtraf.NumPost+".jpeg", msg, chatID)
			if err != nil {
				errorLog.Printf("ошибка отправки фото: %v", err)
				errSend = true
			}
			os.Remove("./" + fmt.Sprint(i) + shtraf.NumPost + ".jpeg")
		}
		if !errSend { //Если при отправки не было ошибок, то добавляем запись
			err = utils.AddEvent(chatID, post, shtraf.DateDiscount)
			if err != nil {
				return countShtaf, err
			}
		}

	}
	infoLog.Printf("Успешная проверка штрафов для %v%v:%v", nomer, region, sts)
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

// sendDiscounts Отправляет инфу о том что заканчивается время оплаты со скидкой
func sendDiscounts(myID int) (err error) {
	mapa, err := utils.GetDiscount()
	if err != nil {
		return
	}
	for key, value := range mapa {
		chatID := value[1]
		t2, err := time.Parse("2006-01-02 15:04:05", value[0]) //Парсим дату
		if err != nil {
			return err
		}
		t1 := time.Now()
		sec := int(t2.Sub(t1)/time.Second - 3*3600) //Вычитаем три часа, так как дата передается без часового пояса
		day := sec / 86400
		hours := (sec - (day * 86400)) / 3600
		minute := (sec - (day*86400 + hours*3600)) / 60
		sesec := (sec - (day*86400 + hours*3600 + minute*60))
		msg := fmt.Sprintf("❗До льготной оплаты по постановлению %v осталось %v дней, %v часов, %v минут, %v секунд\nОплата со скидкой до %v", key, day, hours, minute, sesec, value[0])
		telegram.SendMessage("Debug "+msg, myID)
		id, _ := strconv.Atoi(chatID)
		telegram.SendMessage(msg, id)
	}
	return
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
