package utils

import (
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

var goodProxyList []string //Пустой список с хотстами прокси
func UpdateProxyList() { //Бесконечно обновляет список с хостами прокси
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	reset := "\033[0m"
	infoLog := log.New(os.Stdout, fmt.Sprint(string(colorGreen), "INFO\t"+reset), log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, fmt.Sprint(string(colorRed), "ERROR\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
	goodProxyList = nil //Чистим список
	var newUrl string
	url := "/ru/proxy-list/?type=h#list"
	var proxylist []string    //Прокси-хосты, со всех страниц
	var proxylistNew []string //Прокси-хосты, с одной страницы
	var err error

	// go func() {
	for {
		// wg.Add(1)
		proxylistNew, newUrl, err = getProxy(url)
		if err != nil {
			errorLog.Println(err)
			// wg.Done()
			break
		}
		proxylist = append(proxylist, proxylistNew...)
		if newUrl == "" {
			// wg.Done()
			break
		}
		url = newUrl
	}
	// }()
	// wg.Wait()
	infoLog.Println("Получили списки прокси-хостов со всех страниц")
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
}
func Proxy() (proxyHost string, err error) {
	log.Println("Ищем проксю")
	for i := 0; i < 5; i++ {
		if len(goodProxyList) == 0 {
			if i == 4 {
				err = fmt.Errorf("пустой список прикси-серверов. Нет живых прокси или проблема с доступностью сайта gibdd")
				return
			}
			log.Println("Пустой список прокси серверов, ждем")
			time.Sleep(1 * time.Minute)
			continue
		}
	}

	for _, host := range goodProxyList { //На всякий случай проверяем доступность прикси из хоррошего списка
		err = checkProxy(host)
		if err != nil {
			log.Printf("Прокси-сервер %v протух", host) //Медленный способ но сохраняем порядок. Нам важен порядоок, так как в начале самые быстрые сервера
			continue
		}
		proxyHost = host
		return
	}
	err = fmt.Errorf("нет доступных прокси-хостов или пролема с сайтом gibdd")
	return
}

//Проверяем доступность прокси сервера
func checkProxy(proxy string) (err error) {
	times := time.Now()
	// colorRed := "\033[31m"
	colorGreen := "\033[32m"
	reset := "\033[0m"
	infoLog := log.New(os.Stdout, fmt.Sprint(string(colorGreen), "INFO\t"+reset), log.Ldate|log.Ltime)
	// errorLog := log.New(os.Stderr, fmt.Sprint(string(colorRed), "ERROR\t"+reset), log.Ldate|log.Ltime|log.Lshortfile)
	infoLog.Printf("Проверяем доступность %v", proxy)
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
		Timeout:   time.Second * 10,
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
	infoLog.Printf("Прокся: %v Боди: %v код: %v %v", proxy, string(body), res.StatusCode, time.Since(times).Seconds())
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
		fmt.Println(newUrl, s.Text())
	})
	log.Println(proxyList)
	return
}

type checkStruct struct {
	RequestTime string `json:"requestTime"`
	Hostname    string `json:"hostname"`
	Message     string `json:"message"`
	Status      int    `json:"status"`
}