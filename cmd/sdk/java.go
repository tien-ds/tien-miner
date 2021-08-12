//go:build java
// +build java

package main

import "C"
import (
	"gitee.com/aifuturewell/gojni/java"
	"github.com/ds/depaas/crypto"
)

func main() {

}

func init() {
	java.OnMainLoad(func(reg java.Register) {
		reg.WithClass("com.ds.Fcrypto").
			BindNative("generateKey", "java.lang.String(java.lang.String,java.lang.String)", crypto.GenerateKey).
			BindNative("recoverOwner", "java.lang.String(java.lang.String)", crypto.RecoverOwner).
			BindNative("getUserAddr", "java.lang.String(java.lang.String)", crypto.GetUserAddr).
			BindNative("deCryptFKey", "java.lang.String(java.lang.String,java.lang.String)", crypto.DeCryptFKey).
			BindNative("verifySigned", "java.lang.String(java.lang.String,java.lang.String)", crypto.VerifySigned).
			Done()
	})
}
