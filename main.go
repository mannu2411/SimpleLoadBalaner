package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends")
	flag.IntVar(&port, "port", 3030, "Port to listen on")
	flag.Parse()
	if len(serverList) == 0 {
		log.Fatal("No backends specified")
	}
	tokens := strings.Split(serverList, ",")
	for _, token := range tokens {
		serverURL, err := url.Parse(token)
		if err != nil {
			log.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Println(serverURL, " ", e.Error())
			retries := GetRetryFromContext(request)
			if retries > 3 {
				select {
				case <-time.After(10 * time.Second):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}
			serverPool.MarkBackendStatus(serverURL, false)
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s attempts tried ", attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))

		}
		serverPool.AddBackend(&Backend{
			URL:          serverURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Println("configures server - " + serverURL.String())
	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}
	go healthCheck()
	log.Println("starting server on port " + strconv.Itoa(port))
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
