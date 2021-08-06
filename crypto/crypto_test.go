package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSig(t *testing.T) {
	//user 0xB05146ED865F0aB592DD763bd84a2191700F3dFb
	//0x88d84800F5b7f514f481913D6d20155631Ba2b31
	//69e422d8287f515a63c2461e73db1792f394f61299fdaa74e1749507ad31685b
	p := GenerateKey("69e422d8287f515a63c2461e73db1792f394f61299fdaa74e1749507ad31685b", "0xC1B174F1bc70172c911Df8A01D0e0D98129C7517")
	fmt.Println(p)
}

func TestBase58(t *testing.T) {
	b58 := base58.Decode("4RXZBEMRS9dbRoYM2WLqmQEg7jQrEWqKvDV1ZxESAEEurx4oSmeSfU4BNUGzepqpCXHHBdSyGTwGipayU3kTLVNmEUsEdT3epPXjbVFyt5HXcN5Y9FPNm9xHJJiqnE1pxQtr9KVWkhSkjg76u15uVbdq3Nd3YtDsUgWy")
	fmt.Println(base64.StdEncoding.EncodeToString(b58))
	//fmt.Println(GenerateKey("001"))
}

func TestVery(t *testing.T) {
	a := "3hWfDDQM5wDovJMWFi7SpK4WC8Wa14CDf6CC4PdXhRa3PQcF9Jj43YkmqRvyjiJESY1VyX4vyo68q4cExWJ1EXD7TZf2GhFBvzuB3wfdsR6xPxHDdXXXD"
	fmt.Println(GetUserAddr(a))
}

func TestGen(t *testing.T) {
	//fmt.Println(fmt.Println(GenerateKey("0001")))
	//fmt.Println(GenerateKey("shangzebei"))
}

func TestKey(t *testing.T) {
	fmt.Println(
		FileKey("4EGX5uGg3mPiLdgW3FoQb5NXR3VpjVxfXLbXJbkENeCUTiXximLnSDWTWUZmP4e2B5QzsXHSfp1jba1TD71xWpnADK2Jn9MmJ8x2LQStzoTH5A",
			"C:/Users/Administrator/Desktop/ipfs-pool/ipfs-pool"))
}

func TestFk(t *testing.T) {
	fmt.Println(DeCryptFKey(
		"f640eb336b88e7259df3dd634e96aa27",
		"3TUKv4yEv1iTCs8yZd1RL354LH5kMBC9VwWWiRTUhG3RAANf189p4SvXqi4Hoqb2SLQ9NXCQP3wqdXd68ppbSYJRYko4MSYrF83hskwft4YTbZoP2HKHA"))
}

func TestRecover(t *testing.T) {
	fmt.Println(RecoverOwner("3hWfDDQM5wDovJMWFi7SpK4WC8Wa14CDf6CC4PdXhRa3PQcF9Jj43YkmqRvyjiJESY1VyX4vyo68q4cExWJ1EXD7TZf2GhFBvzuB3wfdsR6xPxHDdXXXD"))
}

//9d89ce7bd0b248b481dee6a01d98ae32f623a3afda77d7677ebc9cc26ea835b1
//0xa0d81025c4314f4e692cf71f9072955d46efa8a5
//51bec94008fd4035b6e33fbd48fdba74e177a8c2ed718cc4106881b248087e5abe972e3cd2822f6b993702f4b224e9bdcc516d7305d0c591542e5f1b19d7910e
func TestLogin(t *testing.T) {
	add := common.HexToAddress("0x28a8507a4dd263358106a7fdfd841159cd979ea6")

	var bytes [32]byte
	copy(bytes[:], add.Bytes())
	copy(bytes[20:], "_login_login")
	fmt.Println(hex.EncodeToString(bytes[:]))
	//Output: a0d81025c4314f4e692cf71f9072955d46efa8a55f6c6f67696e5f6c6f67696e

	pri, _ := crypto.HexToECDSA("3ef73c3896ff7de06d7b39233f8358ea23f56655b29443b4f8a85b44fa6f5bf2")
	fmt.Println(crypto.PubkeyToAddress(pri.PublicKey).String())
	//Output: 0xa0D81025C4314f4E692cF71f9072955D46Efa8A5

	sig, _ := crypto.Sign(Padding(bytes[:]), pri)
	fmt.Println(hex.EncodeToString(sig))
	//Output: 9ab560ce6b26517795fbd74dce999e2f64f73c06169cf6d27226a4a1876a396e624121628459f1fb4c6bcef91d7f278692c8775c61a8dd15f8926a23bd6d8cd601

	fmt.Println(VerifySigned("dda253b1f793abd53413809d5b77364861bb9fd1aee3d540e8782ada2cc5175f1d7ff06f2ff36a130d27ec2bcc3b6e77b71d4c8be40f8d661573cd40fd2c879b01",
		"92fdb6e13c9cca1290031e28c4852337e924039b5F6C6F67696E5F6C6F67696E"))
}
