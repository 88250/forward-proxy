package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/parnurzeal/gorequest"
)

var logger *gulu.Logger

func init() {
	rand.Seed(time.Now().Unix())

	gulu.Log.SetLevel("info")
	logger = gulu.Log.NewLogger(os.Stdout)
}

func handler(w http.ResponseWriter, r *http.Request) {
	result := gulu.Ret.NewResult()
	if "POST" != r.Method {
		result.Code = -1
		result.Msg = "invalid method [" + r.Method + "]"

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = -1

		return
	}

	destURL := args["url"].(string)
	if _, e := url.ParseRequestURI(destURL); nil != e {
		result.Code = -1
		result.Msg = "invalid [url]"

		return
	}

	method := strings.ToUpper(args["method"].(string))

	started := time.Now()

	request := gorequest.New().CustomMethod(method, destURL).Timeout(10*time.Second)
	headers := args["headers"].([]interface{})
	for _, pair := range headers {
		for k, v := range pair.(map[string]interface{}) {
			request.Header.Set(k, fmt.Sprintf("%s", v))
		}
	}

	contentType := args["contentType"]
	if nil != contentType && "" != contentType {
		request.Header.Set("Content-Type", contentType.(string))
	}

	if "POST" == method {
		request.SendString(args["payload"].(string))
	}

	response, bytes, errors := request.EndBytes()
	if nil != errors {
		logger.Infof("request url [%s] failed: %v", destURL, errors)
		result.Code = -1
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
	logger.Infof("elapsed [%.1fs], length [%d], req [url=%s, headers=%s, content-type=%s, body=%s], status [%d], body [%s]",
		duration.Seconds(), len(responseDataBytes), data["url"], headers, contentType, args["payload"], data["status"], shortBody)
}

func main() {
	http.HandleFunc("/", handler)
	logger.Info("Start serving on port 8888")
	http.ListenAndServe("127.0.0.1:8888", nil)
}
