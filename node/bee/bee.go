package bee

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ds/depaas/protocol"

	"github.com/ds/depaas/utils"

	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

func beeClient(url string) ([]byte, error) {
	client := &http.Client{}

	reader := bytes.NewReader([]byte{})

	//logrus.Debug("httpPostClient url:", url, ",query:", string(query))
	request, err := http.NewRequest("GET", url, reader)
	defer request.Body.Close()

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(request)
	if err != nil {
		logrus.Error("httpPostClient err:", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	logrus.Info("beeClient respBytes:", string(respBytes), ",err:", err)
	return respBytes, err
}

type ChequebookStruct struct {
	Address    string `json:"address"`
	Chequebook []struct {
		Peer         string `json:"peer"`
		LastReceived struct {
			Beneficiary string `json:"beneficiary"`
			Chequebook  string `json:"chequebook"`
			Payout      int64  `json:"payout"`
		} `json:"lastReceived"`
		LastSent struct {
			Beneficiary string `json:"beneficiary"`
			Chequebook  string `json:"chequebook"`
			Payout      int64  `json:"payout"`
		} `json:"lastSent"`
	} `json:"chequebook"`
}

type BeeStruct struct {
	protocol.MsgType
	Address      string `json:"address"`
	Peer         string `json:"peer"`
	TotalBalance int64  `json:"totalBalance"`
	Chequebook   []struct {
		Peer         string `json:"peer"`
		LastReceived struct {
			Beneficiary string `json:"beneficiary"`
			Chequebook  string `json:"chequebook"`
			Payout      int64  `json:"payout"`
		} `json:"lastReceived"`
		LastSent struct {
			Beneficiary string `json:"beneficiary"`
			Chequebook  string `json:"chequebook"`
			Payout      int64  `json:"payout"`
		} `json:"lastSent"`
	} `json:"chequebook"`
}

func BeeInfo(id string, peerId string) map[string]interface{} {
	if !utils.IsPortOpen(":1635") {
		logrus.Warn("bzz not start!")
		return nil
	}
	ret := make(map[string]interface{})

	ret["id"] = id
	ret["type"] = 40
	ret["peerID"] = peerId

	//return wifi name
	mac := utils.GetMacWifi()
	if len(mac) >= 4 {
		wifiName := "DS" + mac[4:]
		ret["name"] = wifiName
	}
	data, err := beeClient("http://localhost:1635/addresses")
	if err == nil {
		value, _ := jsonparser.GetString(data, "ethereum")
		ret["address"] = value
		value, _ = jsonparser.GetString(data, "overlay")
		ret["peer"] = value
	} else {
		ret["address"] = ""
		ret["peer"] = ""
	}
	data, err = beeClient("http://localhost:1635/chequebook/balance")
	if err == nil {
		value, _ := jsonparser.GetInt(data, "totalBalance")
		ret["totalBalance"] = value
	} else {
		ret["totalBalance"] = 0
	}
	var chequebook ChequebookStruct
	data, err = beeClient("http://localhost:1635/chequebook/cheque")
	if err == nil {
		err = json.Unmarshal(data, &chequebook)
	}
	if err == nil {
		ret["chequebook"] = chequebook.Chequebook
	} else {
		ret["chequebook"] = make([]interface{}, 0)
	}

	return ret
}

func MakeERC20TransferData(toAddress string, amount *big.Int) ([]byte, error) {
	methodId := crypto.Keccak256([]byte("transfer(address,uint256)"))
	var data []byte
	data = append(data, methodId[:4]...)
	paddedAddress := common.LeftPadBytes(common.HexToAddress(toAddress).Bytes(), 32)
	data = append(data, paddedAddress...)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	data = append(data, paddedAmount...)
	return data, nil
}

func rawERC20Tx(chainId *big.Int, priKey *ecdsa.PrivateKey, to string, nonce uint64, gasPrice *big.Int, gasLimit uint64, amount string) (*types.Transaction, error) {
	value := big.NewInt(0)
	value.SetString(amount, 10)

	data, err := MakeERC20TransferData(to, value)

	contract := common.HexToAddress("0x19062190b1925b5b6689d7073fdfc8c2976ef8cb")
	if chainId.Int64() != 1 {
		contract = common.HexToAddress("0x2ac3c1d3e24b45c6c310534bc2dd84b5ed576335")
	}

	tx := types.NewTransaction(nonce, contract, big.NewInt(0), gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), priKey)

	return signedTx, err
	//err = client.SendTransaction(context, signedTx)
	//return signedTx.Hash().String(), err
}

func KeysToPrivateKey(path string) (string, error) {
	/*
		fks := filekeystore.New(path)
		privateKeyECDSA, _, err := fks.Key("swarm", "1")
		if err != nil {
			logrus.Debug("KeysToPrivateKey err:", err.Error())
			return "", err
		}

		address := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
		logrus.Debug("KeysToPrivateKey address: ", address)

		priKey := "0x" + hex.EncodeToString(crypto.FromECDSA(privateKeyECDSA))
		logrus.Debug("KeysToPrivateKey priKey: ", priKey)
		return priKey, nil

		keyjson, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Debug("KeysToPrivateKey err:", err)
			return "", err
		}

		k, err := keystore.DecryptKey(keyjson, "1")
		if err != nil {
			logrus.Debug("KeysToPrivateKey DecryptKey err:", err)
			return "", err
		}

		logrus.Debug("KeysToPrivateKey key:", hex.EncodeToString(crypto.FromECDSA(k.PrivateKey)))
		return "0x" + hex.EncodeToString(crypto.FromECDSA(k.PrivateKey)), nil*/
	return GetBeeKey("/var/lib/bee/keys/swarm.key", "1")
	//return "0x0ad52e725b52f295fb23f7163f43b6e48b7fdd98d20e7b299c9fed892fe2ca9a", nil
}

func BeeTransfer(url string, params []protocol.BeeParamArgs) (string, error) {

	client, err := ethclient.Dial(url)
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	privateKey, err := KeysToPrivateKey("/var/lib/bee/keys")
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	priKey, err := crypto.HexToECDSA(privateKey[2:])
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	from := crypto.PubkeyToAddress(priKey.PublicKey)
	logrus.Debug("transfer from:", from)
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	context := context.Background()
	chainID, err := client.NetworkID(context)
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	nonce, err := client.PendingNonceAt(context, from)
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	gasPrice, err := client.SuggestGasPrice(context)
	if err != nil {
		logrus.Debug("transfer err:", err)
		return "", nil
	}

	ret := make(map[string]interface{})
	for i, v := range params {
		//logrus.Debug(index, "\t",value)
		amount := big.NewInt(0)
		amount.SetString(v.Amount, 10)

		signedTx, _ := rawERC20Tx(chainID, priKey, v.To, nonce+uint64(i), gasPrice, 76918, v.Amount)
		err = client.SendTransaction(context, signedTx)

		logrus.Debug("transfer signedTx.Hash().String():", signedTx.Hash().String(), ",err:", err)

		if err != nil {
			return "", err
		}

		ret[v.To] = signedTx.Hash().String()
	}

	retdata, _ := json.Marshal(ret)
	return string(retdata), nil
}
