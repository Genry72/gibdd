package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//СheckRegNum Проверяем валидность введенных данных на сайте гибдд
func СheckRegNum(nomer, region, sts, proxyHost string) (err error) {
	//Задаем прокси
	log.Printf("Используем проксю %v", proxyHost)
	proxyStr := "http://" + proxyHost
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return
	}
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
	m := ShtrafStrukt{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга боди запроса на получение списка штрафов: %v Боди: %v", err, string(body))
		return
	}
	if m.Code != 200 {
		err = fmt.Errorf(m.Message)
		return
	}
	return
}

type ShtrafStrukt struct {
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
