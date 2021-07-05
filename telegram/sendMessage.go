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

func SendMessage(message string, chatid int) {
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err := fmt.Errorf("не задан токен")
		log.Println(err)
		return
	}
	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	method := "POST"
	payload := strings.NewReader("chat_id=" + fmt.Sprint(chatid) + "&text=" + message)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		log.Println(err)
		return
	}
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка Парсинга боди ответа на отправку сообщения %v", string(body))
		log.Println(err)
		return
	}
	if !t.Ok {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		log.Println(err)
		return
	}
}

//SendPhoto отправка фото
func SendPhoto(photoName, message string, chatid int) {
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err := fmt.Errorf("не задан токен")
		log.Println(err)
		return
	}
	url := "https://api.telegram.org/bot" + token + "/sendPhoto"
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("chat_id", fmt.Sprint(chatid))
	file, errFile2 := os.Open("./" + photoName)
	if errFile2 != nil {
		log.Println(errFile2)
		return
	}
	defer file.Close()
	part2, errFile2 := writer.CreateFormFile("photo", filepath.Base("./"+photoName))
	if errFile2 != nil {
		log.Println(errFile2)
		return
	}
	_, errFile2 = io.Copy(part2, file)
	if errFile2 != nil {
		log.Println(errFile2)
		return
	}
	_ = writer.WriteField("caption", message)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		log.Println(err)
		return
	}
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка Парсинга боди ответа на отправку сообщения %v", string(body))
		log.Println(err)
		return
	}
	if !t.Ok {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		log.Println(err)
		return
	}
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
