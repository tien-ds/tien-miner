package bee

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
)

const (
	keyHeaderKDF = "scrypt"
	keyVersion   = 3

	scryptN     = 1 << 15
	scryptR     = 8
	scryptP     = 1
	scryptDKLen = 32
)

// This format is compatible with Ethereum JSON v3 key file format.
type encryptedKey struct {
	Address string    `json:"address"`
	Crypto  keyCripto `json:"crypto"`
	Version int       `json:"version"`
}

type keyCripto struct {
	Cipher       string       `json:"cipher"`
	CipherText   string       `json:"ciphertext"`
	CipherParams cipherParams `json:"cipherparams"`
	KDF          string       `json:"kdf"`
	KDFParams    kdfParams    `json:"kdfparams"`
	MAC          string       `json:"mac"`
}

type cipherParams struct {
	IV string `json:"iv"`
}

type kdfParams struct {
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	DKLen int    `json:"dklen"`
	Salt  string `json:"salt"`
}

func GetBeeKey(name, password string) (string, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read private key: %w", err)
	}

	pk, err := decryptKey(data, password)
	if err != nil {
		return "", err
	}

	priKey := "0x" + hex.EncodeToString(crypto.FromECDSA(pk))
	return priKey, nil
}

func decryptKey(data []byte, password string) (*ecdsa.PrivateKey, error) {
	var k encryptedKey
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, err
	}
	if k.Version != keyVersion {
		return nil, fmt.Errorf("unsupported key version: %v", k.Version)
	}
	d, err := decryptData(k.Crypto, password)
	if err != nil {
		return nil, err
	}
	return DecodeSecp256k1PrivateKey(d)
}

func DecodeSecp256k1PrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	if l := len(data); l != btcec.PrivKeyBytesLen {
		return nil, fmt.Errorf("secp256k1 data size %d expected %d", l, btcec.PrivKeyBytesLen)
	}
	privk, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	return (*ecdsa.PrivateKey)(privk), nil
}

func decryptData(v keyCripto, password string) ([]byte, error) {
	if v.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("unsupported cipher: %v", v.Cipher)
	}

	mac, err := hex.DecodeString(v.MAC)
	if err != nil {
		return nil, fmt.Errorf("hex decode mac: %s", err)
	}
	cipherText, err := hex.DecodeString(v.CipherText)
	if err != nil {
		return nil, fmt.Errorf("hex decode cipher text: %s", err)
	}
	derivedKey, err := getKDFKey(v, []byte(password))
	if err != nil {
		return nil, err
	}
	calculatedMAC := sha3.Sum256(append(derivedKey[16:32], cipherText...))
	if !bytes.Equal(calculatedMAC[:], mac) {
		return nil, errors.New("invalid password")
	}

	iv, err := hex.DecodeString(v.CipherParams.IV)
	if err != nil {
		return nil, fmt.Errorf("hex decode IV cipher parameter: %s", err)
	}
	data, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, nil
}

func getKDFKey(v keyCripto, password []byte) ([]byte, error) {
	if v.KDF != keyHeaderKDF {
		return nil, fmt.Errorf("unsupported KDF: %s", v.KDF)
	}
	salt, err := hex.DecodeString(v.KDFParams.Salt)
	if err != nil {
		return nil, fmt.Errorf("hex decode salt: %s", err)
	}
	return scrypt.Key(
		password,
		salt,
		v.KDFParams.N,
		v.KDFParams.R,
		v.KDFParams.P,
		v.KDFParams.DKLen,
	)
}
