package main

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type Reader struct {
	rd       io.Reader
	callback func(uint int)
}

func NewReader(reader io.Reader, callback func(uint int)) *Reader {
	return &Reader{
		rd:       reader,
		callback: callback,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	read, err := r.rd.Read(p)
	r.callback(read)
	return read, err
}

func UploadFile(filePath, token string) {
	url := "http://39.99.129.137:22334/api/pool/upload2"

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

	//progress bar
	bar := progressbar.DefaultBytes(
		fileSize,
		"UPLOADFILE",
	)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, NewReader(file, func(uint int) {
		bar.Add(uint)
	}))
	req.Header.Add("token", token)
	if err != nil {
		logrus.Error(err)
		return
	}
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
