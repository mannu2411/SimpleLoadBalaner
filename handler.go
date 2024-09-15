package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var serverPool ServerPool

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("attempts exceeded")
		http.Error(w, "attempts exceeded", http.StatusServiceUnavailable)
		return
	}
	peer := serverPool.GetNextPeer()
	if peer == nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service Not Available", http.StatusServiceUnavailable)
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Error connecting to backend:", err)
		return false
	}
	defer conn.Close()
	return true
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func healthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Health Check Started")
			serverPool.HealthCheck()
			log.Println("Health Check Finished")
		}
	}
}
