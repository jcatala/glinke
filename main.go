package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/PuerkitoBio/goquery"

)


var v bool

const version = "0.0.1"

func main() {

	verbose := flag.Bool("v", false, "To be verbose")
	concurrency := flag.Int("threads", 10, "Threads to use, default: 10")
	t := flag.Int("t",1000,"timeout in milliseconds")

	flag.Parse()
	input := make(chan string)
	output := make(chan string)
	v = *verbose
	if *verbose{
		fmt.Println("Verbose mode")
		fmt.Printf("Version: %s\n", version)
	}
	// time duration for net.Dialer
	timeout := time.Duration(*t * 100000000)

	var tr = &http.Transport{
		MaxIdleConns: 30,
		IdleConnTimeout: time.Second,
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{Timeout: timeout, KeepAlive: time.Second}).DialContext,
	}
	re := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10{
			return http.ErrUseLastResponse
		}
		return nil
	}
	client := &http.Client{Transport: tr, CheckRedirect: re, Timeout: timeout}

	// Input worker
	var inputWG sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		inputWG.Add(1)
		go func() {
			for url := range input{
				// Since the results are already probe, just go to find the respective links
				if *verbose {
					log.Println("Got the following url" , url)
				}
				for _,link := range grabLinks(client, url ){
					output <- link
				}
			}
			inputWG.Done()
		}()
	}

	// output worker
	var outputWG sync.WaitGroup
	outputWG.Add(1)
	go func(){
		for url := range output {
			fmt.Println(url)
		}
		outputWG.Done()
	}()

	// Close channel output when the input workers are done
	go func() {
		inputWG.Wait()
		close(output)
	}()


	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan(){
		target := strings.ToLower(sc.Text())
		if *verbose{
			log.Println("domain", target)
		}
		input <- target
	}
	close(input)
	outputWG.Wait()

}


func grabLinks(client *http.Client, url string) []string{
	// if the url does not terminate on /, add it
	if url[len(url)-1] != '/' {
		url = url + "/"
	}
	var ret []string
	// logic of grabbing links :D
	res, err := client.Get(url)
	if err != nil {
		if v {
			log.Println("Error: ", err)
		}
		return nil
	}
	defer res.Body.Close()
	// Load the document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil{
		if v{
			log.Println("Error on creating goquery reader: ", err)
		}
		return nil
	}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		dsrc, _ := s.Attr("data-src")
		if src != "" {
			if !strings.HasPrefix(src, "http"){
				src = url + src
			}
			ret = append(ret, src)
		}
		if dsrc != "" {
			if !strings.HasPrefix(dsrc, "http"){
				dsrc = url + src
			}
			ret = append(ret, dsrc)
		}
	})

	return ret
}