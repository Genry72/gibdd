package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var colorRed = "\033[31m"
// var colorGreen = "\033[32m"
// var colorYellow = "\033[33m"
var reset = "\033[0m"
// var infoLog = log.New(os.Stdout, fmt.Sprint(string(colorGreen), "INFO\t"+reset), log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, fmt.Sprint(string(colorRed), "ERROR\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
// var warnLog = log.New(os.Stdout, fmt.Sprint(string(colorYellow), "WARN\t"+reset), log.Ldate|log.Ltime)

func SendMessage(message string, chatid int) (err error) {
	token := os.Getenv("telegaGibddToken")
	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	method := "POST"
	payload := strings.NewReader("chat_id=" + fmt.Sprint(chatid) + "&text=" + message)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		errorLog.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		errorLog.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		errorLog.Println(err)
		return
	}
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка Парсинга боди ответа на отправку сообщения %v", string(body))
		errorLog.Println(err)
		return
	}
	if !t.Ok {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		errorLog.Println(err)
		return
	}
	return
}

//SendPhoto отправка фото
func SendPhoto(photoName, message string, chatid int) (err error) {
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err = fmt.Errorf("не задан токен")
		errorLog.Println(err)
		return
	}
	url := "https://api.telegram.org/bot" + token + "/sendPhoto"
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("chat_id", fmt.Sprint(chatid))
	file, errFile2 := os.Open("./" + photoName)
	if errFile2 != nil {
		errorLog.Println(errFile2)
		return errFile2
	}
	defer file.Close()
	part2, errFile2 := writer.CreateFormFile("photo", filepath.Base("./"+photoName))
	if errFile2 != nil {
		errorLog.Println(errFile2)
		return errFile2
	}
	_, errFile2 = io.Copy(part2, file)
	if errFile2 != nil {
		errorLog.Println(errFile2)
		return errFile2
	}
	_ = writer.WriteField("caption", message)
	err = writer.Close()
	if err != nil {
		errorLog.Println(err)
		return errFile2
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		errorLog.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		errorLog.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errorLog.Println(err)
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		errorLog.Println(err)
		return
	}
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка Парсинга боди ответа на отправку сообщения %v", string(body))
		errorLog.Println(err)
		return
	}
	if !t.Ok {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		errorLog.Println(err)
		return
	}
	return
}

type sendMsgTelegaStruct struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int    `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}
