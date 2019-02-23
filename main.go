package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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
	for k, v := range r.Header {
		request.Header.Set(k, fmt.Sprintf("%s", v))
	}

	response, bytes, errors := request.EndBytes()
	if nil != errors {
		logger.Infof("get url [%s] failed: %v", destURL, errors)
		result.Code = CodeErr
		result.Msg = "internal error"

		return
	}

	responseBody := string(bytes)
	responseData := map[string]interface{}{
		"url":         destURL,
		"status":      response.StatusCode,
		"contentType": response.Header.Get("content-type"),
		"body":        responseBody,
		"headers":     response.Header,
	}

	responseDataBytes, e := json.Marshal(responseData)
	if nil != e {
		logger.Errorf("marshal original response failed %#v", e)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	key := args["key"].(string)
	if "" != key {
		responseDataBytes = AESEncrypt(key, responseDataBytes)
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
		duration.Seconds(), len(responseDataBytes), responseData["url"], responseData["status"], shortBody)
}

func main() {
	http.HandleFunc("/", handler)
	logger.Info("Start serving on port 8888")
	http.ListenAndServe(":8888", nil)
}

func AESEncrypt(key string, data []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil
	}
	ecb := cipher.NewCBCEncrypter(block, []byte("RandomInitVector"))
	content := data
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	return crypted
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AESDecrypt(key string, crypt []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil
	}
	ecb := cipher.NewCBCDecrypter(block, []byte("RandomInitVector"))
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return PKCS5Trimming(decrypted)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
