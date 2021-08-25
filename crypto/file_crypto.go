package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/ds/depaas/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"io"
	"os"
)

const Admin = "0x88d84800F5b7f514f481913D6d20155631Ba2b31"

func FileKey(key string, file string) string {
	f, e := os.Open(file)
	if e != nil {
		panic(e)
	}
	return BufferKey(key, f)
}

func BufferKey(key string, r io.Reader) string {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		panic(err)
	}
	c := h.Sum(nil)
	return FKeyWithMd5(key, c)
}

func FKeyWithMd5(pkey string, md5 []byte) string {
	rs, e := utils.Des3Encrypt([]byte(pkey), Resize24(md5))
	if e != nil {
		panic(e)
	}
	return base58.Encode(rs)
}

func Resize24(b []byte) []byte {
	if len(b) != 16 {
		panic(fmt.Sprintf("%d!=16", len(b)))
	}
	var a []byte
	a = append(a, b[:16]...)
	a = append(a, b[:8]...)
	return a
}

// DeCryptFKey 34d25cbb29f2d2af 172122b1337f043a023c39401a4aa6f3
func DeCryptFKey(md5, token string) string {
	md5bytes, e := hex.DecodeString(md5)
	if e != nil {
		panic(e)
	}
	rs, e := utils.Des3Decrypt(base58.Decode(token), Resize24(md5bytes))
	if e != nil {
		panic(e)
	}
	return string(rs)
}

// GenerateKey 0xC1B174F1bc70172c911Df8A01D0e0D98129C7517
//"69e422d8287f515a63c2461e73db1792f394f61299fdaa74e1749507ad31685b"
func GenerateKey(ownerPrivateKey, addr string) string {
	bAddr := common.HexToAddress(addr)
	if len(bAddr) != 20 {
		panic("len error must < 15")
	}
	prefix := base58.Encode(Padding(bAddr[:]))
	pri, _ := crypto.HexToECDSA(ownerPrivateKey)
	//32
	var bytes [32]byte
	copy(bytes[:], bAddr[:])
	sig, _ := crypto.Sign(Padding(bytes[:]), pri)
	return prefix + base58.Encode(sig)
}

func Padding(b []byte) []byte {
	//for i := 0; i < len(b); i++ {
	//	if b[i] == 0 {
	//		b[i] = '0'
	//	}
	//}
	return b
}

func RecoverOwner(pkey string) string {
	addr := base58.Decode(pkey[:28])
	signed := base58.Decode(pkey[28:])
	var bytes [32]byte
	copy(bytes[:], addr)
	pub, e := crypto.SigToPub(Padding(bytes[:]), signed)
	if e != nil {
		return e.Error()
	}
	return crypto.PubkeyToAddress(*pub).String()
}

func VerifyKey(pubKey, key string) bool {
	return RecoverOwner(key) == pubKey
}

func GetUserAddr(key string) string {
	if VerifyKey(Admin, key) {
		orig := base58.Decode(key[:28])
		return common.BytesToAddress(orig).String()
	} else {
		return "GetUserId error"
	}
}

// VerifySigned raw 32
func VerifySigned(signed, raw string) string {
	sBytes, err := hex.DecodeString(signed)
	if err != nil {
		panic(err)
	}
	sRaw, err := hex.DecodeString(raw)
	if err != nil {
		panic(err)
	}
	pub, e := crypto.SigToPub(sRaw, sBytes)
	if e != nil {
		panic(e)
	}
	return crypto.PubkeyToAddress(*pub).String()
}
