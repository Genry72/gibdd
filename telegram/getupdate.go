package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

//Getupdate возвращает офсет для последующего получения обновлений
func GetOfset() (offset int, err error) {
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err = fmt.Errorf("не задан токен")
		return offset, err
	}
	url := "https://api.telegram.org/bot" + token + "/getUpdates"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return offset, err
	}
	res, err := client.Do(req)
	if err != nil {
		return offset, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return offset, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка получения офсета %v: %v", res.Status, err)
		return offset, err
	}
	m := GetupdateStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение офсета: %v Боди: %v", err, string(body))
		return offset, err
	}
	if len(m.Result) == 0 {
		return offset, err
	}
	return m.Result[0].UpdateID, err
}

func Getupdate(offset int) (message, sender string, chatID int, newOffset int, username string, err error) {
	log.Printf("Получаем сообщения из телеги, офсет %v", offset)
	token := os.Getenv("telegaGibddToken")
	if token == "" {
		err = fmt.Errorf("не задан токен")
		return message, sender, chatID, newOffset, username, err
	}
	url := "https://api.telegram.org/bot" + token + "/getUpdates?timeout=50&offset=" + fmt.Sprintf("%v", offset)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return message, sender, chatID, newOffset, username, err
	}
	res, err := client.Do(req)
	if err != nil {
		return message, sender, chatID, newOffset, username, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return message, sender, chatID, newOffset, username, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("ошибка получения офсета %v: %v", res.Status, err)
		return message, sender, chatID, newOffset, username, err
	}
	m := GetupdateStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение офсета: %v Боди: %v", err, string(body))
		return message, sender, chatID, newOffset, username, err
	}
	if len(m.Result) == 0 { //Таймаут по долгому запросу
		log.Printf("Сработал таймаут Боди:%v", string(body))
		return message, sender, chatID, offset, username, err
	}
	if m.Result[0].Message.Text == "" { //Тип сообщения не текст
		log.Println("Тип не текст")
		return message, sender, chatID, offset + 1, username, err
	}
	return m.Result[0].Message.Text, m.Result[0].Message.From.FirstName + " " + m.Result[0].Message.From.LastName, m.Result[0].Message.Chat.ID, offset + 1, m.Result[0].Message.From.Username, err
}

type GetupdateStruct struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
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
		} `json:"message"`
	} `json:"result"`
}
