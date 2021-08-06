package utils

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	fmt.Println(GenRand())
}

func TestDir(t *testing.T) {

}

func TestFile(t *testing.T) {
	fmt.Println(Md5file("C:/Users/Administrator/Downloads/app-release-v2.4.apk"))
}

func TestConnect(t *testing.T) {
	timeout, err := net.DialTimeout("tcp", "123.100.236.30:18080", time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Println(timeout)
}
