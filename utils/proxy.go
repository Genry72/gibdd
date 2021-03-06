package utils

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var reset = "\033[0m"
var infoLog = log.New(os.Stdout, fmt.Sprint(string(colorGreen), "INFO\t"+reset), log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, fmt.Sprint(string(colorRed), "ERROR\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
var warnLog = log.New(os.Stdout, fmt.Sprint(string(colorYellow), "WARN\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
var goodProxyList []string //Пустой список с хотстами прокси

func UpdateProxyList() (err error) { //Бесконечно обновляет список с хостами прокси
	goodProxyList = nil //Чистим список
	var newUrl string
	url := "/ru/proxy-list/?type=h#list"
	var proxylist []string    //Прокси-хосты, со всех страниц
	var proxylistNew []string //Прокси-хосты, с одной страницы

	for { //Собираем грязный список с сайта https://hidemy.name/
		// wg.Add(1)
		proxylistNew, newUrl, err = getProxy(url)
		if err != nil {
			errorLog.Println(err)
			break
		}
		proxylist = append(proxylist, proxylistNew...)
		if newUrl == "" {
			break
		}
		url = newUrl
	}
	proxylistNew, err = getProxyProxyscrape() //Забираем грязный список с сайта https://proxyscrape.com/
	if err != nil {
		errorLog.Println(err)
	}
	proxylist = append(proxylist, proxylistNew...)
	infoLog.Println("Получили списки прокси-хостов со всех страниц")
	infoLog.Println("Проверяем доступность и составляем список годных прокси-серверов")
	//Проверяем доступность прокси хостов из общего списка, формируя при этом новый
	for _, proxy := range proxylist {
		go func(proxy string) {
			err = checkProxy(proxy)
			if err != nil {
				return
			}
			goodProxyList = append(goodProxyList, proxy)
		}(proxy)
	}
	return nil //Зануляем, так как пробрасывается из функции getProxyProxyscrape
}

//Proxy просто проверяет готовность списка хостов
func Proxy() (proxylist []string, err error) {
	for i := 0; i < 5; i++ {
		if len(goodProxyList) == 0 {
			if i == 4 {
				err = fmt.Errorf("пустой список прикси-серверов. Нет живых прокси или проблема с доступностью сайта gibdd")
				return
			}
			warnLog.Println("Пустой список прокси серверов, ждем")
			time.Sleep(1 * time.Minute)
			continue
		}
	}
	proxylist = goodProxyList
	return
}

//Проверяем доступность прокси сервера
func checkProxy(proxy string) (err error) {
	// times := time.Now()
	//Задаем прокси
	proxyStr := "http://" + proxy
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
		Timeout:   30 * time.Second,
	}
	url := "https://check.gibdd.ru/proxy/check/fines"
	method := "POST"

	payload := strings.NewReader("regnum=%D0%90777%D0%90%D0%90&regreg=777&stsnum=7777777777")
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
		err = fmt.Errorf("код не 200 %v %v", res.Status, string(body))
	}
	t := checkStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = fmt.Errorf("ошибка парсинга %v Боди:%v", err, string(body))
		return
	}
	if t.Hostname != "check.gibdd.ru" {
		err = fmt.Errorf("хрень в ответе %v", string(body))
		return
	}
	// infoLog.Printf("Годный прокси-сервер %v Время ответа %v", proxy, time.Since(times).Seconds())
	return
}

func getProxy(url string) (proxyList []string, newUrl string, err error) {
	res, err := http.Get("https://hidemy.name" + url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = fmt.Errorf("status code error: %v", res.Status)
		return
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	// Find the review items
	// doc.Find(".drag_list .drag_element .slider--wide slider__cut-hover clearfix .slide__item .post-img picture .picture-holder").Each(func(i int, s *goquery.Selection) {
	for g := 0; g < 100; g++ { //Не более 100 результатов на странице
		doc.Find("body > div.wrap > div.services_proxylist.services > div > div.table_block > table").Each(func(i int, s *goquery.Selection) {
			hosts := s.Find("tr:nth-child(" + fmt.Sprintf("%v", g) + ") > td:nth-child(1)")
			host := strings.Replace(hosts.Text(), "IP адрес", "", -1)
			if host == "" {
				return
			}
			ports := s.Find("tr:nth-child(" + fmt.Sprintf("%v", g) + ") > td:nth-child(2)")
			port := strings.Replace(ports.Text(), "Порт", "", -1) //Удаляем лишний текст
			proxyList = append(proxyList, host+":"+port)
		})
	}
	doc.Find(".next_array").Each(func(i int, s *goquery.Selection) {
		qq := s.Find("a")
		newUrl, _ = qq.Attr("href")
		// fmt.Println(newUrl, s.Text())
	})
	// log.Println(proxyList)
	return
}

func getProxyProxyscrape() (proxyList []string, err error) {
	url := "https://api.proxyscrape.com/v2/?request=getproxies&protocol=http&timeout=10000&country=RU&ssl=all&anonymity=all&simplified=true"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(string(body))) //Построчно читаем что прилетело в канал
	for scanner.Scan() {
		proxyList = append(proxyList, scanner.Text())
	}
	return
}

type checkStruct struct {
	RequestTime string `json:"requestTime"`
	Hostname    string `json:"hostname"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}
