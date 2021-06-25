package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	nomer := "Х752ТТ"
	region := "152"
	sts := "9933143213"
	err := checkShtraf(nomer, region, sts)
	if err != nil {
		err = fmt.Errorf("ошибка при получении штрафов: %v", err)
		log.Println(err)
	}

}

//checkShtraf Функция проверки штрафов
func checkShtraf(nomer, region, sts string) (err error) {
	url := "https://check.gibdd.ru/proxy/check/fines"
	method := "POST"
	payload := strings.NewReader("regnum=" + nomer + "&regreg=" + region + "&stsnum=" + sts)
	client := &http.Client{}
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
	m := shtrafStrukt{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение списка штрафов: %v Боди: %v", err, string(body))
		return err
	}
	var post string
	var divid int
	if m.Code != 200 {
		err = fmt.Errorf("выполнение запроса %v завершиблось ошибкой %v: %v", url, m.Message, string(body))
		return err
	}
	cafapPicsToken := m.CafapPicsToken
	for _, shtraf := range m.Data {
		dateNarush := shtraf.DateDecis
		fmt.Printf("Дата нарушения %v\n", dateNarush)
		post = shtraf.NumPost
		divid = shtraf.Division
		err = linkImage(post, nomer+region, fmt.Sprintf("%v", divid), cafapPicsToken)
		if err != nil {
			err = fmt.Errorf("ошибка получения картинки со штрафом: %v", err)
			return err
		}
	}
	// fmt.Println(string(body))
	return err
}

//linkImage Получаем ссылку на картинку
func linkImage(post, regnum, divid, cafapPicsToken string) (err error) {
	url := "https://check.gibdd.ru/proxy/check/fines/pics"
	method := "POST"
	payload := strings.NewReader("post=" + post + "&regnum=" + regnum + "&divid=" + divid + "&cafapPicsToken=" + cafapPicsToken)
	client := &http.Client{}
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
	link := fmt.Sprintf("%v", m.Photos[0].Base64Value)
	err = base64toJpg("./screen.jpeg", link)
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
	fmt.Println("Jpg file", filepath, "created")
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
