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
			return err
		}
		defer f.Close()
		c, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		if strings.Contains(string(c), "idle") {
			return err
		}
		log.Printf("Ждем запуск диска %v", string(c))
		time.Sleep(5 * time.Second)
	}
	err = fmt.Errorf("не дождались старта диска")
	return
}
