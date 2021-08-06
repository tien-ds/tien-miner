package config

import (
	"encoding/base64"
	"github.com/ds/depaas/utils"
	"github.com/loomnetwork/go-loom"
	"github.com/sirupsen/logrus"
	"github.com/tien-ds/contract-miner/miner"
)

const (
	ChainKey = "ChainKey"
	BindAddr = "BindAddr"
	Config   = "CONFIG"
	IP       = "IP"
)

func GetChainKey() string {
	priKey, err := GetConfigKey(ChainKey)
	if err != nil {
		key := miner.GenKey()
		err := SetConfigKey(ChainKey, key.Pri)
		if err != nil {
			logrus.Errorf("not save loom private key err %s", err)
		}
		logrus.Debugf("GetChainKey key.Pri %s", key.Pri)
		return key.Pri
	}
	logrus.Tracef("GetChainKey priKey %s", priKey)
	return priKey
}

func GetChainAddress() string {
	privateKey, _ := base64.StdEncoding.DecodeString(GetChainKey())
	publicKey := make([]byte, 32)
	copy(publicKey, privateKey[32:])

	addr := loom.LocalAddressFromPublicKey(publicKey[:])
	address := base64.StdEncoding.EncodeToString(addr)
	logrus.Debug("ChainAddress address:", address)
	return address
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
