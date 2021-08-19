package service

import (
	"bytes"
	"encoding/json"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/persistence"
	"github.com/ds/depaas/pools"
	"github.com/ds/depaas/protocol"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var subscribes map[int]string

func toDB() {
	j, _ := json.Marshal(subscribes)
	persistence.GetOrigDB().Put([]byte("SUBSCRIBES_DB"), j)
}

func InitFormDB() {
	v, e := persistence.GetOrigDB().Get([]byte("SUBSCRIBES_DB"))
	logrus.Info(string(v), e)
	if e == nil && v != nil {
		err := json.Unmarshal(v, &subscribes)
		if err != nil {
			logrus.Error(err)
		}
	}
	if subscribes == nil {
		subscribes = make(map[int]string)
	}
}

func SubscribesInit() {
	InitFormDB()
	api.POST(func(a struct {
		URL  string `json:"url"`
		Type int    `json:"type"`
	}) interface{} {
		var k = false
		if _, b := subscribes[a.Type]; b {
			k = true
		}
		subscribes[a.Type] = a.URL
		toDB()
		return map[string]bool{
			"override": k,
		}
	}, "/api/pool/subscribe")

	//binds
	pools.MsPool().Register(protocol.MINER_BIND_RESP.String(), func(id string, f interface{}) {
		v, _ := json.Marshal(f)
		do(id, v)
	})

	//info
	pools.MsPool().Register(protocol.MINFO.String(), func(id string, f interface{}) {
		kk := f.(*protocol.InfoType)
		kk.Time = time.Now().Unix()
		v, _ := json.Marshal(kk)
		do(id, v)
	})

	//
	pools.MsPool().Register(protocol.ONLINE_OR_OFFLINE.String(), func(id string, f interface{}) {
		v, _ := json.Marshal(f)
		do(id, v)
	})

	//
	pools.MsPool().Register(protocol.MESSAGE.String(), func(id string, f interface{}) {
		v, _ := json.Marshal(f)
		do(id, v)
	})

	pools.MsPool().Register(protocol.CHIA_INFO.String(), func(id string, f interface{}) {

		sum := func(info *protocol.ChiaMinerInfo) *big.Int {
			s := big.NewInt(0)
			for _, plot := range info.Plots {
				s = s.Add(s, big.NewInt(plot.Size))
			}
			return big.NewInt(0).Div(s, big.NewInt(1024*1024*1024))
		}

		push := func(cm *protocol.ChiaMinerInfo) {
			v, _ := json.Marshal(f)
			do(id, v)
		}

		kk := f.(*protocol.ChiaMinerInfo)

		if kk.ID == "7894531" { //bind
			push(kk)
			return
		}

		var old protocol.ChiaMinerInfo
		pct := persistence.GetStoreDB().Get(kk.PeerID, reflect.TypeOf(protocol.ChiaMinerInfo{}))

		//store last
		kk.Time = time.Now().Unix()
		err := persistence.GetStoreDB().Store(kk.PeerID, kk)
		if err != nil {
			logrus.Error(err)
		}

		if pct == nil {
			push(kk)
			return
		}

		if v, b := pct.(*protocol.ChiaMinerInfo); !b {
			push(kk)
			return
		} else {
			old = *v
		}

		newSum := sum(kk)
		oldSum := sum(&old)
		logrus.Infof("old=%s G old time %d,new=%s G", oldSum, old.Time, newSum)
		if newSum.Cmp(oldSum) != 0 {
			push(kk)
		}
	})

	pools.MsPool().Register(protocol.BEE_INFO.String(), func(id string, f interface{}) {
		v, _ := json.Marshal(f)
		do(id, v)
	})
}

func do(id string, body []byte) {
	if v, e := strconv.Atoi(id); e == nil {
		if url, e1 := subscribes[v]; e1 {
			go post(url, body)
		}
	}
}

func post(url string, body []byte) bool {
	logrus.Infof("push url %s body %s", url, string(body))
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Warn(err)
		return false
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		logrus.Warn(err)
		return false
	}
	defer res.Body.Close()
	rep, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn(err)
		return false
	}

	b, err := strconv.ParseBool(string(rep))
	if err != nil {
		logrus.Warn(err)
		return false
	}
	return b
}
