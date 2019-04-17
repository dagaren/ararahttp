package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

func main() {
	config := clientConfiguration{}

	flag.StringVar(&config.Url, "Url", "", "Url")
	flag.IntVar(&config.NumRequests, "Requests", 1, "Number of requests")
	flag.BoolVar(&config.UseHttp2, "Http2", false, "Use http2")
	flag.IntVar(&config.Concurrency, "Concurrency", 1, "Number of concurrent requests")
	flag.Parse()

	fmt.Println(" - Url:", config.Url)
	fmt.Println(" - Num requests: ", config.NumRequests)

	client := &http.Client{}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	if config.UseHttp2 == true {
		var t = &http2.Transport{
			TLSClientConfig: tlsConfig,
		}

		client.Transport = t
	} else {
		var t = &http.Transport{
			TLSClientConfig:     tlsConfig,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 10,
		}

		client.Transport = t
		fmt.Println("Transport disable keepalives: ", t.DisableKeepAlives)
		fmt.Println("Transport max idle conns: ", t.MaxIdleConns)
		fmt.Println("Transport max idle conns per host: ", t.MaxIdleConnsPerHost)
		//fmt.Println("Transport max accounts per host: ", t.MaxConnsPerHost)
		fmt.Printf("%+v\n", t)
	}

	channel := make(chan string, 2)
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	for i := 0; i < config.Concurrency; i++ {
		GetBodyFromChannel(client, channel, &wg)
	}

	start := time.Now()

	for i := 0; i < config.NumRequests; i++ {
		url := fmt.Sprintf("%s/%d", config.Url, i)
		channel <- url
	}
	close(channel)
	wg.Wait()

	elapsed := time.Since(start)

	fmt.Println(" - Execution time", elapsed)

	fmt.Println("Normal termination of the program")
}

func GetBodyFromChannel(client *http.Client, channel chan string, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for url := range channel {
			_ = GetBody(client, url)
		}
	}()
}

func GetBody(client *http.Client, url string) string {
	response, err := client.Get(url)

	if err != nil {
		fmt.Println("Failed get: %s", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed reading response body: %s", err)
	}
	// fmt.Printf(
	//	"Got response %d: %s %s\n",
	//	response.StatusCode, response.Proto, string(body))

	//fmt.Println(string(body))
	return string(body)
}

type clientConfiguration struct {
	Url         string
	NumRequests int
	UseHttp2    bool
	Concurrency int
}
