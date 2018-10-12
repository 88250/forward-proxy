package main

import (
	"log"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	response, bytes, errors := gorequest.New().Get(url).Timeout(5 * time.Second).Retry(2, 3*time.Second).EndBytes()
	if nil != errors {
		log.Printf("get url [%s] failed: %v", url, errors)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(response.StatusCode)
		w.Write(bytes)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Start serving on port 8888")
	http.ListenAndServe(":8888", nil)
}
