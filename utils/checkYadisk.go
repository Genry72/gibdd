package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

//CheckYadisk Проверка готовности яндекс диска
func CheckYadisk() (err error) {
	for i := 0; i < 60; i++ {
		f, err := os.Open("./yadisk/.sync/status")
		if err != nil {
			log.Printf("Ждем запуск диска %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer f.Close()
		c, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("Ждем запуск диска %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if strings.Contains(string(c), "idle") {
			infoLog.Println("YandexDisk запущен")
			return err
		}
		log.Printf("Ждем запуск диска %v", string(c))
		time.Sleep(5 * time.Second)
	}
	err = fmt.Errorf("не дождались старта диска")
	return
}
