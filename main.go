package main

import (
	"github.com/jcatala/glinke/scrapit"
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
				for _,link := range scrapit.GrabLinks(client, url,v ){
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

