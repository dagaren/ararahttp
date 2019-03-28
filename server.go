package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	config := serverConfiguration{}
	requestConfig := requestConfiguration{}

	flag.IntVar(&config.Port, "Port", 8080, "Local port")
	flag.IntVar(&config.AdminPort, "AdminPort", 8081, "Admin Local Port")
	flag.BoolVar(&config.DisableKeepAlive, "DisableKeepAlive", false, "Disable keep alive")
	flag.BoolVar(&config.Ssl, "Secure", false, "Secure server")
	flag.StringVar(&config.SslCert, "SslCert", "", "Secure server")
	flag.StringVar(&config.SslKey, "SslKey", "", "Secure server")
	flag.IntVar(&requestConfig.Duration, "Duration", 32, "Request duration")
	flag.Parse()

	adminHandler := http.NewServeMux()
	mockHandler := http.NewServeMux()

	mockHandler.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		fmt.Printf("Request to mock received %s. Host: %s, URL: %s\n", time.Now().Format(time.RFC3339), request.Host, request.URL.Path)

		time.Sleep(time.Duration(requestConfig.Duration) * time.Millisecond)
		//response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(requestConfig.Response))
	})

	adminHandler.HandleFunc("/config", func(response http.ResponseWriter, request *http.Request) {
		fmt.Printf("Request to admin received %s. Method: %s, Host: %s, URL: %s\n", request.Method, time.Now().Format(time.RFC3339), request.Host, request.URL.Path)

		switch request.Method {
		case "POST":
			contentType := request.Header.Get("Content-Type")
			fmt.Println("Content type", contentType)
			newConfig := requestConfiguration{}

			if contentType != "application/json" {
				http.Error(response, "Invalid content type", 405)
			} else {
				decoder := json.NewDecoder(request.Body)
				err := decoder.Decode(&newConfig)

				requestConfig = newConfig

				if err != nil {
					http.Error(response, "Invalid content", 405)
				} else {
					fmt.Println("Redirecting...")
					http.Redirect(response, request, "/config", 303)
				}
			}
		case "GET":
			response.Header().Add("Content-Type", "application/json")
			json.NewEncoder(response).Encode(requestConfig)
		default:
			http.Error(response, "Invalid request method", 405)
		}
	})

	fmt.Println("Requests duration", requestConfig.Duration, "milliseconds")

	go func() {
		fmt.Println("Starting admin server on port", config.AdminPort)
		fmt.Println("Disable keep alive", config.DisableKeepAlive)
		adminServer := &http.Server{Addr: fmt.Sprintf(":%d", config.AdminPort), Handler: adminHandler}
		adminServer.SetKeepAlivesEnabled(!config.DisableKeepAlive)
		if config.Ssl == false {
			if err := adminServer.ListenAndServe(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := adminServer.ListenAndServeTLS(config.SslCert, config.SslKey); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	}()

	fmt.Println("Starting mock server on port", config.Port)
	mockServer := &http.Server{Addr: fmt.Sprintf(":%d", config.Port), Handler: mockHandler}
	mockServer.SetKeepAlivesEnabled(!config.DisableKeepAlive)
	if config.Ssl == false {
		if err := mockServer.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := mockServer.ListenAndServeTLS(config.SslCert, config.SslKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Normal termination of the program")
}

type serverConfiguration struct {
	Port             int
	AdminPort        int
	DisableKeepAlive bool
	Ssl              bool
	SslKey           string
	SslCert          string
}

type requestConfiguration struct {
	Duration int
	Response string
}
