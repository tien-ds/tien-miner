package main

import (
	"bytes"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func UploadFile(filePath, token string) {
	url := "http://39.99.129.137:22334/api/pool/uploadFile"
	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)

	//Get file size
	fileSize := int64(0)
	if s, e := os.Stat(filePath); e == nil {
		fileSize = s.Size()
	}

	//Open file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	part1, err := writer.CreateFormFile("file", filepath.Base(filePath))

	//progress bar
	bar := progressbar.DefaultBytes(
		fileSize,
		"UPLOADFILE",
	)
	_, err = io.Copy(io.MultiWriter(part1, bar), file)
	if err != nil {
		logrus.Error(err)
		return
	}
	_ = writer.WriteField("token", token)
	err = writer.Close()
	if err != nil {
		logrus.Error(err)
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		logrus.Error(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Println(string(body))

}
