package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

var (
	tempDir = filepath.Join(os.TempDir(), "ds")
)

func init() {
	if _, error := os.Stat(tempDir); os.IsNotExist(error) {
		os.Mkdir(tempDir, 0777)
	}
}

func GetTmpSize(tempName string) int64 {
	if f, e := os.Stat(tempName); e == nil {
		return f.Size()
	} else {
		return 0
	}
}

func GetTempDir() string {
	return tempDir
}

func WriteToTemp(offset int64, reader io.Reader, tempName string) (int64, string, error) {
	sName := filepath.Join(tempDir, tempName)
	logrus.Debugf("dir %s", sName)
	var f *os.File
	var fErr error
	if offset > 0 && Exist(sName) {
		f, fErr = os.OpenFile(sName, os.O_RDWR, 0666)
	} else {
		f, fErr = os.Create(sName)
	}
	defer f.Close()
	if fErr != nil {
		panic(fErr)
	}
	if offset > 0 {
		tmpSize := GetTmpSize(sName)
		logrus.Infof("has cached %d", tmpSize)
		if offset > tmpSize {
			panic(fmt.Sprintf("has cached %d but %d more %d", tmpSize, offset, offset-tmpSize))
		}
		fmt.Println(f.Seek(offset, io.SeekCurrent))
	}
	w, err := io.Copy(f, reader)
	if err != nil {
		logrus.Errorf("name %s has cache %d", sName, w)
		return 0, "", err
	}
	return w, sName, nil
}

func Md5File(f io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Md5file(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return Md5File(f)
}

func Md5String(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}

func Exist(name string) bool {
	if _, e := os.Stat(name); e == nil {
		return true
	} else {
		return false
	}
}
