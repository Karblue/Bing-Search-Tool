package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

//BingData is Bing Search Data
type BingData struct {
	URL   string
	Title string
	Owned string
}

var (
	iProxyURL    string // Global Proxy URL,default is empty string
	iType        string // Type is C(C seg) OR S(Single), default is S
	iHostOrIP    string // Search Host Or IP Address
	iKeyword     string // Use Keyword Search Data,default is empty string , if not empty, override type and host , force use keyword search
	iStopCount   int    // Fetch Stop Count,default 0 is no limited
	iWorkerCount int    // worker count, default is 10

)

var (
	globalProxy *url.URL
)

func httpGet(url string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(globalProxy),
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; ) AppleWebKit/547.36 (KHTML, like Gecko) Chrome/76.0.3729.169 Safari/547.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respByte), nil

}

//get next page bing search data
func nextBingSearchData(url string) ([]BingData, string) {
	var bMatchData []BingData
	var sNextURL string

	content, err := httpGet(url)
	if err != nil {
		fmt.Println("[*] Http Get Bing Error : ", err)
		return nil, sNextURL
	}

	if strings.Index(content, "Ref A:") > -1 {
		fmt.Println("[*] Please use IP address not in Chinese Mainland to visit bing.com")
		fmt.Println(content)
		return nil, sNextURL
	}
	regMain := regexp.MustCompile("<main.+?</main>")
	mainContent := regMain.FindString(content)
	reg := regexp.MustCompile("<h2><a href=\"(.+?)\" h=\".+?\">(.+?)</a></h2>")
	result := reg.FindAllStringSubmatch(mainContent, -1) //search url link
	if len(result) > 0 {
		bMatchData = make([]BingData, 0)
		for _, data := range result {
			if len(data[1]) > 0 && len(data[2]) > 0 {
				bBingData := BingData{
					URL:   data[1],
					Title: data[2],
				}
				bMatchData = append(bMatchData, bBingData)
			}
		}
	} else {
		return nil, sNextURL
	}

	regNextPage := regexp.MustCompile("<a class=\"sb_pagN sb_pagN_bp b_widePag sb_bp \".+ href=\"(.+?)\" h=\".+?\"><div class=\"sw_next\"")
	nextPageResult := regNextPage.FindAllStringSubmatch(content, -1)
	if len(nextPageResult) > 0 && len(nextPageResult[0]) > 0 {
		nextURL := nextPageResult[0][1]
		sNextURL = "https://www.bing.com" + strings.Replace(nextURL, "&amp;", "&", -1)
	}
	//fmt.Println(nextPageResult)

	return bMatchData, sNextURL
}

//get first bing search data
func getBingSearchData(keyword string) []BingData {
	keywordValue := url.Values{}
	keywordValue.Add("q", keyword)
	url := fmt.Sprintf("https://www.bing.com/search?%s&first=1&FORM=PERE", keywordValue.Encode())
	var lastNextURL string
	var resultCount int
	mapURL := make(map[string]int)
	allBingData := make([]BingData, 0)
	fmt.Println(keyword)
	for {
		if len(url) == 0 {
			break
		}

		if iStopCount != 0 && resultCount >= iStopCount {
			break
		}
		data, nextURL := nextBingSearchData(url)

		if nextURL == lastNextURL || mapURL[nextURL] == 1 {
			break // already last page
		}

		//fmt.Printf("Page Count : %d ,Next URL :%s \n", len(data), nextURL)
		for _, item := range data {
			fmt.Printf("URL:%s,Title:\"%s\" \n", item.URL, item.Title)
		}
		allBingData = append(allBingData, data...)
		lastNextURL = nextURL
		url = nextURL
		resultCount += len(data)
		mapURL[nextURL] = 1
	}
	fmt.Println()
	return allBingData
}

// set http proxy
func setProxy(addr string) {
	urli := url.URL{}
	globalProxy, _ = urli.Parse(addr)
}

//get host dns ip address
func getHostIPAddress(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	return addrs[0], nil

}

// get c segment ip address
func getCSegIPAddress(ip string) ([]string, error) {
	reg := regexp.MustCompile("((25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))\\.){3}(25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))")
	if !reg.MatchString(ip) {
		return nil, errors.New("IP Address not validate")
	}
	ipaddrs := make([]string, 0)
	lastIndex := strings.LastIndex(ip, ".")
	cIPAddress := ip[:lastIndex]

	for i := 1; i < 255; i++ {
		ipaddrs = append(ipaddrs, fmt.Sprintf("%s.%d", cIPAddress, i))
	}
	return ipaddrs, nil

}
func saveBingDatasHTML(filename string, dataMap map[string][]BingData) error {
	html := fmt.Sprintf("<html>%s<body>%s</body></html>", getCSS(), getBingDatasHTML(dataMap))
	return ioutil.WriteFile(filename, []byte(html), os.ModePerm)
}

func getBingDatasHTML(dataMap map[string][]BingData) string {
	var html string
	if len(dataMap) == 0 {
		return "Not Data"
	}

	for keyword, datas := range dataMap {
		html += fmt.Sprintf("<ol class=\"rounded-list\"><li><a>%s</a></li><ol>", keyword)
		for _, data := range datas {
			html += fmt.Sprintf("<li><a target='_blank' href='%s'>%s</a></li>", data.URL, data.Title)
		}
		html += "</ol></ol><br/>"
	}

	return html
}

func getCSS() string {
	return `<style>
	body{
		margin: 40px auto;
		width: 500px;
	}

	ol{
		counter-reset: li;
		list-style: none;
		*list-style: decimal;
		font: 15px 'trebuchet MS', 'lucida sans';
		padding: 0;
		margin-bottom: 4em;
		text-shadow: 0 1px 0 rgba(255,255,255,.5);
	}

	ol ol{
		margin: 0 0 0 2em;
	}

	/* -------------------------------------- */			

	.rounded-list a{
		position: relative;
		display: block;
		padding: .4em .4em .4em 2em;
		*padding: .4em;
		margin: .5em 0;
		background: #ddd;
		color: #444;
		text-decoration: none;
		-moz-border-radius: .3em;
		-webkit-border-radius: .3em;
		border-radius: .3em;
		-webkit-transition: all .3s ease-out;
		-moz-transition: all .3s ease-out;
		-ms-transition: all .3s ease-out;
		-o-transition: all .3s ease-out;
		transition: all .3s ease-out;	
	}

	.rounded-list a:hover{
		background: #eee;
	}

	.rounded-list a:hover:before{
		-moz-transform: rotate(360deg);
		-webkit-transform: rotate(360deg);
		-moz-transform: rotate(360deg);
		-ms-transform: rotate(360deg);
		-o-transform: rotate(360deg);
		transform: rotate(360deg);	
	}

	.rounded-list a:before{
		content: counter(li);
		counter-increment: li;
		position: absolute;	
		left: -1.3em;
		top: 50%;
		margin-top: -1.3em;
		background: #87ceeb;
		height: 2em;
		width: 2em;
		line-height: 2em;
		border: .3em solid #fff;
		text-align: center;
		font-weight: bold;
		-moz-border-radius: 2em;
		-webkit-border-radius: 2em;
		border-radius: 2em;
		-webkit-transition: all .3s ease-out;
		-moz-transition: all .3s ease-out;
		-ms-transition: all .3s ease-out;
		-o-transition: all .3s ease-out;
		transition: all .3s ease-out;
	}
</style>`
}

func getCsegBingData(ips []string) map[string][]BingData {
	var waitGroup = sync.WaitGroup{}
	limitChan := make(chan int, iWorkerCount)
	lock := make(chan int, 1)
	allMapData := make(map[string][]BingData)
	waitGroup.Add(len(ips))
	for _, ip := range ips {
		go func(searchIP string, dataMap map[string][]BingData) {
			limitChan <- 1
			allData := getBingSearchData("ip:" + searchIP)
			<-limitChan
			if len(allData) > 0 {
				lock <- 1
				dataMap[searchIP] = allData
				<-lock
			}
			waitGroup.Done()

		}(ip, allMapData)
	}
	waitGroup.Wait()
	return allMapData
}

func main() {

	flag.StringVar(&iProxyURL, "p", "", "proxy url,like http://127.0.0.1:1087")
	flag.StringVar(&iType, "t", "S", "search type, value is C or S(single host)")
	flag.StringVar(&iHostOrIP, "h", "", "search host or ip address")
	flag.StringVar(&iKeyword, "k", "", "force use keyword search")
	flag.IntVar(&iStopCount, "s", 0, "if result-count > x will stop, default 0 is not limited")
	flag.IntVar(&iWorkerCount, "w", 10, "worker count")
	flag.Usage = func() {
		fmt.Println("bing search tools v.20190524 \nUsage: bing [-p proxy-url] [-t C/S] [-h hostOrip] [-k keyword] [-s stop-count] [-w worker thread]\n\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(iKeyword) == 0 && len(iHostOrIP) == 0 {
		flag.Usage()
	} else {
		if len(iProxyURL) > 0 {
			setProxy(iProxyURL)
		}
		allMapData := make(map[string][]BingData)
		if len(iKeyword) > 0 { // force use keyword
			allMapData[iKeyword] = getBingSearchData(iKeyword)
		} else {
			ipaddr, err := getHostIPAddress(iHostOrIP)
			if err != nil {
				fmt.Printf("[*] Can't Resolve Host %s,Error :%s \n ", iHostOrIP, err)
				return
			}

			fmt.Printf("[*] Host %s IP Address is %s\n", iHostOrIP, ipaddr)

			if strings.ToLower(iType) == "c" {
				ipaddrs, err := getCSegIPAddress(ipaddr)
				if err != nil {
					fmt.Printf("[*] Can't Resolve C Segment IPv4 address %s,Error :%s \n ", iHostOrIP, err)
					return
				}
				allMapData = getCsegBingData(ipaddrs)
			} else {
				allMapData[ipaddr] = getBingSearchData("ip:" + ipaddr)
			}
		}

		saveBingDatasHTML("result.html", allMapData)

	}
}
