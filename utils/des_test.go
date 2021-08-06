package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCBC(t *testing.T) {
	//c8928ea0c629c95ee280cf59
	//s, e := hex.DecodeString("c8928ea0c629c95ee280cf59aa0538a3")
	//resize24 := Resize24(s)
	//fmt.Println(hex.EncodeToString(resize24))
	b, e := Des3Encrypt([]byte("hello"), []byte("c8928ea0c629c95ee280cf59"))
	fmt.Println(hex.EncodeToString(b), e)

	//Output
	//bc4a212be9ed8da5 <nil>
	de, _ := Des3Decrypt(b, []byte("c8928ea0c629c95ee280cf59"))
	fmt.Println(string(de))

	//Output
	//hello
}

func Test3Des(t *testing.T) {
	//
	key := "3TUKv4yEv1iTCs8yZd1RL354LH5kMBC9VwWWiRTUhG3RAANf189p4SvXqi4Hoqb2SLQ9NXCQP3wqdXd68ppbSYJRYko4MSYrF83hskwft4YTbZoP2HKHA"
	sb, _ := hex.DecodeString("f640eb336b88e7259df3dd634e96aa27")
	resize24 := Resize24(sb)
	fmt.Println(resize24)
	// Output: [246 64 235 51 107 136 231 37 157 243 221 99 78 150 170 39 246 64 235 51 107 136 231 37]
	desEn, _ := Des3Encrypt([]byte(key), resize24)
	fmt.Println(hex.EncodeToString(desEn))

	// Output: cb426d9537cab8a72718d7add6bb78a089eb15c7d3601f2ebfd8154f0b12433c85833d080746400f72d12b9fc20a271ce38fd9d89890b40a1fffe23a72b5c60c325a19254c0850a6a9b3b570ef717ae156f3bfa23365bc8132c9a8a441beb95552bd4dbbea795d90e18b1fc5a46991fe77e0dc0dac0c1fcb
}

func Resize24(b []byte) []byte {
	if len(b) != 16 {
		panic("error")
	}
	var a []byte
	a = append(a, b[:16]...)
	a = append(a, b[:8]...)
	return a
}
