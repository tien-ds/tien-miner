//+build share

package main

import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
)

//export sign
func sign(rawHash *C.char, privateHash *C.char) *C.char {
	dataBytes := C.GoString(rawHash)
	fmt.Println("raw ", dataBytes)
	fmt.Println("pri ", dataBytes)
	priBytes := C.GoString(privateHash)
	pri, err := crypto.HexToECDSA(priBytes)
	if err != nil {
		panic(err)
	}
	//fmt.Println(crypto.PubkeyToAddress(pri.PublicKey).String())
	//Output: 0xa0D81025C4314f4E692cF71f9072955D46Efa8A5
	bytes, err := hex.DecodeString(dataBytes)
	if err != nil {
		panic(err)
	}
	sig, err := crypto.Sign(bytes, pri)
	if err != nil {
		panic(err)
	}
	return C.CString(hex.EncodeToString(sig))
}

func main() {

}
