package config

import (
	"encoding/base64"
	"github.com/ds/depaas/utils"
	"github.com/sirupsen/logrus"
	"github.com/tien-ds/contract-miner/miner"
)

const (
	ChainKey = "ChainKey"
	BindAddr = "BindAddr"
	Config   = "CONFIG"
	IP       = "IP"
)

func GetChainPrivateKey() string {
	priKey, err := GetConfigKey(ChainKey)
	if err != nil {
		key := miner.GenKey()
		err := SetConfigKey(ChainKey, key.Pri)
		if err != nil {
			logrus.Errorf("not save loom private key err %s", err)
		}
		logrus.Debugf("GetChainPrivateKey key.Pri %s", key.Pri)
		return key.Pri
	}
	logrus.Tracef("GetChainPrivateKey priKey %s", priKey)
	return priKey
}

func GetChainAddress() string {
	addr := miner.PrivateKeyToAddr(GetChainPrivateKey())
	logrus.Debug("ChainAddress ", addr)
	return addr
}

func GetBindAddr() string {
	value, err := GetConfigKey(BindAddr)
	if err != nil {
		return ""
	}
	return value
}

func SetBindAddr(addr string) error {
	return SetConfigKey(BindAddr, addr)
}

func GetDSConfig() string {
	value, err := GetConfigKey(Config)
	if err != nil {
		return ""
	}
	decodeString, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return ""
	}
	ds := string(utils.AesDecryptCBC(decodeString, utils.AesPasswd()))
	return ds
}

func SetDSConfig(buf string) error {
	msgBytes := utils.AesEncryptCBC([]byte(buf), utils.AesPasswd())
	bEncode := base64.StdEncoding.EncodeToString(msgBytes)
	return SetConfigKey(Config, bEncode)
}
