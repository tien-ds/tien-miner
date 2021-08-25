package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/ds/depaas/ipds"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitee.com/fast_api/api/def"
	"github.com/ds/depaas/crypto"
	"github.com/ds/depaas/service"
	"github.com/ds/depaas/utils"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/sirupsen/logrus"
)

/**
*
*               100 bytes                                                                                                  -
*
*   +-----------------------------------+
*   |              file                 |
*   +-----------------------------------+
*   ---------------------|
*      60 bytes had done |
*                        |---------------
*                        |40 bytes remain
*                        |
*                        |
*                     offset
*
 */

func getValue(kvs map[string][]string, key string) *string {
	token, b := kvs[key]
	if !b {
		return nil
	}
	if token == nil || len(token) == 0 {
		return nil
	}
	return &token[0]

}

//support upload file continuous
func uploadFile1(header def.Header, reader multipart.Reader) interface{} {
	token := header.Get("token")
	offset := header.Get("offset")
	name := header.Get("name")
	size := header.Get("size")
	off := 0
	if offset != "" {
		mOff, e := strconv.Atoi(offset)
		if e != nil {
			panic(e)
		}
		off = mOff
	}
	p, err := reader.NextPart()
	if err != nil {
		panic(err)
	}
	dSize, tempName, ferr := utils.WriteToTemp(int64(off), p, utils.Md5String(name))
	if ferr != nil {
		return errors.New(fmt.Sprintf("%d", dSize))
	}
	totalSize, e := strconv.Atoi(size)
	if e != nil {
		panic(e)
	}
	if int64(totalSize) != utils.GetTmpSize(tempName) {
		panic("size !=")
	}
	defer os.Remove(tempName)
	return toIpfs(tempName, token)
}

func toIpfs(tempName, sToken string) interface{} {
	//has temp file
	md5, err := utils.Md5file(tempName)
	if err != nil {
		return err
	}

	pkey := crypto.DeCryptFKey(md5, sToken)

	if !crypto.VerifyKey(crypto.Admin, pkey) {
		return errors.New("no trust CA")
	}

	//TODO GET User
	user := crypto.GetUserAddr(pkey)
	logrus.Infof("user is %s", user)

	tf, err := os.Open(tempName)
	if err != nil {
		return err
	}

	pHash, err := ipds.GetApi().Unixfs().Add(context.Background(), files.NewReaderFile(tf), checkSize)
	if err == nil {
		go service.ReceiveCid(pHash.Cid())
	} else {
		logrus.Error(err)
		return err
	}
	logrus.Debugf("ok %s", pHash.Cid().String())
	return map[string]string{"result": pHash.Cid().String()}
}

func uploadFile2(req http.Request) interface{} {
	token := req.Header.Get("token")
	fileName := utils.Md5String(time.Now().String())
	_, tempName, _ := utils.WriteToTemp(0, req.Body, fileName)
	return toIpfs(tempName, token)
}

func uploadFile(reader multipart.Reader) interface{} {
	f, err := reader.ReadForm(100000)
	if err != nil {
		return err
	}
	sToken := getValue(f.Value, "token")
	if sToken == nil {
		return errors.New("no token")
	}

	offset := getValue(f.Value, "offset")

	var tempName string
	if f, b := f.File["file"]; b && len(f) != 0 {
		fileHeader := f[0]
		if fileHeader.Size == 0 {
			return errors.New("file size is 0")
		}
		fileName := utils.Md5String(fileHeader.Filename)
		openF, e := fileHeader.Open()
		if e != nil {
			return e
		}
		off := 0
		if offset != nil {
			off, e = strconv.Atoi(*offset)
			if e != nil {
				panic(e)
			}
		}
		dSize, name, _ := utils.WriteToTemp(int64(off), openF, fileName)
		tempName = name
		if dSize == fileHeader.Size {
			defer os.Remove(name)
		} else {
			return def.NewErrorCode(fmt.Sprintf("%d", utils.GetTmpSize(fileName)), 1)
		}
	} else {
		return errors.New("not file context")
	}

	return toIpfs(tempName, *sToken)
}
