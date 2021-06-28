package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func SendMessage(message string, chatid int) (err error) {
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err = fmt.Errorf("не задан токен")
		return err
	}
	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	method := "POST"
	payload := strings.NewReader("chat_id=" + fmt.Sprint(chatid) + "&text=" + message)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		return err
	}
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка Парсинга боди ответа на отправку сообщения %v", string(body))
		return err
	}
	if !t.Ok {
		err = fmt.Errorf("ошибка отправки сообщения в телеграм %v: %v", res.Status, string(body))
		return err
	}
	return err
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
