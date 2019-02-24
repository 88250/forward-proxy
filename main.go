package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/parnurzeal/gorequest"
)

var logger *Logger

func init() {
	rand.Seed(time.Now().Unix())

	SetLevel("info")
	logger = NewLogger(os.Stdout)
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := NewResult()
	if "POST" != r.Method {
		result.Code = CodeErr
		result.Msg = "invalid method [" + r.Method + "]"

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = CodeErr

		return
	}

	destURL := args["url"].(string)
	if _, e := url.ParseRequestURI(destURL); nil != e {
		result.Code = CodeErr
		result.Msg = "invalid [url]"

		return
	}

	started := time.Now()

	request := gorequest.New().Get(destURL).Timeout(10*time.Second).Retry(2, time.Second)
	headers := args["headers"].([]interface{})
	for _, pair := range headers {
		for k, v := range pair.(map[string]interface{}) {
			request.Header.Set(k, fmt.Sprintf("%s", v))
		}
	}

	response, bytes, errors := request.EndBytes()
	if nil != errors {
		logger.Infof("get url [%s] failed: %v", destURL, errors)
		result.Code = CodeErr
		result.Msg = "internal error"

		return
	}

	responseBody := string(bytes)
	data := map[string]interface{}{
		"url":         destURL,
		"status":      response.StatusCode,
		"contentType": response.Header.Get("content-type"),
		"body":        responseBody,
		"headers":     response.Header,
	}
	result.Data = data

	responseDataBytes, e := json.Marshal(result)
	if nil != e {
		logger.Errorf("marshal original response failed %#v", e)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	w.Write(responseDataBytes)

	duration := time.Now().Sub(started)
	shortBody := ""
	if 64 > len(responseBody) {
		shortBody = responseBody
	} else {
		shortBody = responseBody[:64]
	}
	logger.Infof("ellapsed [%.1fs], length [%d], URL [%s], status [%d], body [%s]",
		duration.Seconds(), len(responseDataBytes), data["url"], data["status"], shortBody)
}

func main() {
	http.HandleFunc("/", handler)
	logger.Info("Start serving on port 8888")
	http.ListenAndServe(":8888", nil)
}
