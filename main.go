package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/parnurzeal/gorequest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	destURL := r.URL.Query().Get("url")
	if _, e := url.ParseRequestURI(destURL); nil != e {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, bytes, errors := gorequest.New().Get(destURL).Timeout(5*time.Second).Retry(2, 3*time.Second).EndBytes()
	if nil != errors {
		log.Printf("get url [%s] failed: %v", destURL, errors)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		header := w.Header()
		for k, v := range response.Header {
			header.Add(k, fmt.Sprintf("%s", v))
		}

		w.Write(bytes)
		w.WriteHeader(response.StatusCode)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Start serving on port 8888")
	http.ListenAndServe(":8888", nil)
}
