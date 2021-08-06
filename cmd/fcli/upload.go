package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func UploadFile(filePath, token string) ([]byte, error) {
	url := "http://127.0.0.1:8099/api/pool/upload"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(filePath)
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(filePath))
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		logrus.Error(errFile1)
		return nil, errFile1
	}
	_ = writer.WriteField("token", token)
	err := writer.Close()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return body, nil
}
