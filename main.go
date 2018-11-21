package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/parnurzeal/gorequest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	started := time.Now()

	destURL := r.URL.Query().Get("url")
	if _, e := url.ParseRequestURI(destURL); nil != e {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	request := gorequest.New().Get(destURL).Timeout(10*time.Second).Retry(2, time.Second)
	for k, v := range r.Header {
		request.Header.Set(k, fmt.Sprintf("%s", v))
	}

	response, bytes, errors := request.EndBytes()
	if nil != errors {
		log.Printf("get url [%s] failed: %v", destURL, errors)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	responseBody := string(bytes)
	responseData := map[string]interface{}{
		"URL":         destURL,
		"Status":      response.StatusCode,
		"ContentType": response.Header.Get("content-type"),
		"Body":        responseBody,
		"Headers":     response.Header,
	}

	header := w.Header()
	for k, v := range response.Header {
		header.Add(k, fmt.Sprintf("%s", v[0]))
	}

	responseDataBytes, e := json.Marshal(responseData)
	if nil != e {
		log.Printf("marshal original response failed %#v", e)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	key := r.URL.Query().Get("key")
	if "" != key {
		responseDataBytes = AESEncrypt(key, responseDataBytes)
	}

	retData := map[string]interface{}{"Data": base64.StdEncoding.EncodeToString(responseDataBytes)}
	retDataBytes, e := json.Marshal(retData)
	if nil != e {
		log.Printf("marshal response failed %#v", e)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	w.Write(retDataBytes)

	duration := time.Now().Sub(started)
	shortBody := ""
	if 64 > len(responseBody) {
		shortBody = responseBody
	} else {
		shortBody = responseBody[:64]
	}
	log.Printf("ellapsed [%.1fs], length [%d], URL [%s], status [%d], %s",
		duration.Seconds(), len(retDataBytes), responseData["URL"], responseData["Status"], shortBody)
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Start serving on port 8888")
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
