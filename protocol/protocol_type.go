package protocol

import (
	"errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

type (
	CMD        int
	MSG        int
	BeeCmdType int
	MinerType  int
)

const (
	PINADD CMD = iota
	PINRM
	CAT
	GET
)

//矿机类型 2家庭矿机  1云矿机 3超级矿机
const (
	MINER_CLOUD MinerType = 1
	MINER_HOME  MinerType = 2
	MINER_SUPER MinerType = 3
)

const (
	BEE_START    BeeCmdType = 1
	BEE_STOP     BeeCmdType = 2
	BEE_TRANSFER BeeCmdType = 3
	DE_TRANSFER  BeeCmdType = 4
)

const (
	HELLO_CODE        MSG = 0
	ID_CODE           MSG = 1
	MINER_PEER        MSG = 2
	AES_ENCRYPT       MSG = 0xae
	CMD_SYSTEM        MSG = 3
	MINFO             MSG = 4
	CMD_SYSTEM_RESP   MSG = 5
	CHIA_INFO_ORIGN   MSG = 6
	CMD_CODE          MSG = 8
	CMD_RESP_SUCCESS  MSG = 9
	CMD_RESP_FAIL     MSG = 10
	ONLINE_OR_OFFLINE MSG = 14
	MINER_BIND_RESP   MSG = 16
	MNEMONIC_CODE     MSG = 20
	MNEMONIC_RESP     MSG = 21
	SELF_UPDATE_CODE  MSG = 22
	SELF_UPDATE_RESP  MSG = 23
	WIFI_KEY_PAIR     MSG = 24
	CHIA_INFO         MSG = 30
	CHIA_INFO_CODE    MSG = 31
	BEE_INFO          MSG = 40
	BEE_CMD_CODE      MSG = 42
	BLOCK_CHECK       MSG = 43
)

func (m MSG) Int() int {
	return int(m)
}

func (m MSG) String() string {
	return strconv.Itoa(m.Int())
}

type MsgType struct {
	ID   string `json:"id"`
	Type MSG    `json:"type"`
}

func SetMsgTypeID(f interface{}, id string) {
	logrus.Tracef("SetMsgTypeID %s", id)
	vs := reflect.ValueOf(f)
	vtye := vs.Type()
	if vtye.Kind() == reflect.Ptr {
		vtye = vtye.Elem()
	}
	if vs.Kind() == reflect.Ptr {
		vs = vs.Elem()
	}
	if vtye == reflect.TypeOf((*MsgType)(nil)).Elem() {
		vs.FieldByName("ID").Set(reflect.ValueOf(id))
	}
}

func GetMsgTypeID(f interface{}) (string, error) {
	vs := reflect.ValueOf(f)
	vtye := vs.Type()
	if vtye.Kind() == reflect.Ptr {
		vtye = vtye.Elem()
	}
	if vs.Kind() == reflect.Ptr {
		vs = vs.Elem()
	}
	if vtye == reflect.TypeOf((*MsgType)(nil)).Elem() {
		return vs.FieldByName("ID").String(), nil
	}
	for i := 0; i < vtye.NumField(); i++ {
		field := vtye.Field(i)
		ftyp := field.Type
		if ftyp.Kind() == reflect.Ptr {
			ftyp = ftyp.Elem()
		}
		if field.Anonymous && ftyp == reflect.TypeOf((*MsgType)(nil)).Elem() {
			return vs.Field(i).Field(0).String(), nil
		} else {
			logrus.Error("is not impl protocol.MsgType")
		}
	}
	return "", errors.New("is not impl protocol.MsgType")
}

// HelloType type 0
type HelloType struct {
	MsgType
	Msg string `json:"msg"`
}

// AesType type 0xae
type AesType struct {
	MsgType
	Msg string `json:"msg"`
}

// Miner type 2
type Miner struct {
	MsgType
	PeerID    string `json:"peerId"`
	Addr      string `json:"addr" db:"index"`
	MinerAddr string `json:"minerAddr"`
}

type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CmderPrikey type 50
type CmderPrikey struct {
	MsgType
	Key     string `json:"key"`
	Address string `json:"address"`
}

// SendEntry type 3
type SendEntry struct {
	MsgType
	Cmd       CMD     `json:"cmd"`
	Version   int     `json:"version"`
	Signature string  `json:"signature"`
	Params    []Param `json:"params"`
	RandId    string  `json:"randId"`
}

// ResultType type 5
type ResultType struct {
	MsgType
	Error  string `json:"error"`
	Result string `json:"result"`
	RandID string `json:"randId"`
}

// InfoType type 4
type InfoType struct {
	MsgType
	PeerId   string `json:"peerId" db:"index"`
	CPUCount int    `json:"cpuCount"`
	//GB
	DiskTotal   float64 `json:"diskTotal"`
	DownSpeed   float64 `json:"downSpeed"`
	UpSpeed     float64 `json:"upSpeed"`
	IpOnLine    bool    `json:"ipOnLine"`
	RAM         int64   `json:"ram"`
	MachineType int     `json:"machineType"`
	TotalIn     int64   `json:"totalIn"`
	TotalOut    int64   `json:"totalOut"`
	Worth       float64 `json:"worth"`
	Time        int64   `json:"time" db:"index"`
	RepoSize    int64   `json:"repoSize"`
	Addr        string  `json:"addr"`
	Ip          string  `json:"ip"`
}

// BlockchainInfo type 6
type BlockchainInfo struct {
	MsgType
	PeerId     string      `json:"peerId"`
	Difficulty int         `json:"difficulty"`
	Height     int         `json:"height"`
	Timestamp  interface{} `json:"timestamp"`
	Balance    int64       `json:"balance"`
	Address    string      `json:"address"`
	State      bool        `json:"state"`
	SyncMode   bool        `json:"sync_mode"`
	Network    string      `json:"network"`
	VDFCount   int         `json:"VDFCount"`
	AllCount   int         `json:"allCount"`
	XCHDay     int         `json:"XCHDay"`
	Space      int64       `json:"space"`
	Size       int64       `json:"size"`
	ChiaID     string      `json:"chiaId"`
}

// Mnemonic type 20,21
type Mnemonic struct {
	Mnemonic string `json:"mnemonic"`
	ID       string `json:"id"`
	Type     int    `json:"type"`
}

// CmderUpdate type 22
type CmderUpdate struct {
	MsgType
	Arm64   string `json:"arm64"`
	Amd64   string `json:"amd64"`
	Version int    `json:"version"`
}

// CmderUpdateRet type 23
type CmderUpdateRet struct {
	MsgType
	Code int `json:"code"` //0，表示不需要升级，1 表示升级成功，2 表示升级中， 3 表示升级出错
}

type PlotStruct struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ChiaMinerInfo type 30
type ChiaMinerInfo struct {
	MsgType
	Plots  []PlotStruct `json:"plots"`
	PeerID string       `json:"peerId"`
	Time   int64
}

// GETChiaMinerInfo type 31
type GETChiaMinerInfo struct {
	MsgType
}

func CmdUpdate(id string, armUrl string, amdUrl string, version int) *CmderUpdate {
	return &CmderUpdate{
		MsgType: MsgType{
			ID:   id,
			Type: SELF_UPDATE_CODE,
		},
		Arm64:   armUrl,
		Amd64:   amdUrl,
		Version: version,
	}
}

// Cmder type 8
type Cmder struct {
	MsgType
	Params []string `json:"params"` //命令行如 ["ls", "-al"]
}

// CmderResult type 9, 10, 9是命令输出，10是命令错误码, ""表示没有错误，错误格式为"error:${msg}"
type CmderResult struct {
	MsgType
	Text string `json:"text"`
}

func HelloOk(id string, ip string) HelloType {
	return HelloType{
		MsgType: MsgType{
			ID:   id,
			Type: HELLO_CODE,
		},
		Msg: ip,
	}
}

func CmdMsg(id string, s string) *Cmder {
	sCmd := strings.Split(s, " ")
	if len(sCmd) == 0 {
		return nil
	}
	return &Cmder{
		MsgType: MsgType{
			ID:   id,
			Type: CMD_CODE,
		},
		Params: sCmd,
	}
}

type Seed struct {
	MsgType
}

func CmdSeed(id string) *Seed {
	return &Seed{
		MsgType: MsgType{
			ID:   id,
			Type: MNEMONIC_CODE,
		},
	}
}

type Wifi struct {
	MsgType
	Wifi   string `json:"wifi"`
	PeerID string `json:"peerId"`
}

// Bee type 40
type Bee struct {
	MsgType
	Address      string `json:"address"`
	PeerID       string `json:"peerId"`
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

func BeeCmd(id string, cmdTye BeeCmdType, ethUrl string, param []BeeParamArgs) *CmderBee {
	c := &CmderBee{
		MsgType: MsgType{
			ID:   id,
			Type: BEE_CMD_CODE,
		},
		Cmd: cmdTye,
		Url: ethUrl,
	}
	if param != nil {
		c.Params = param
	}
	return c
}

type BeeParamArgs struct {
	Amount string `json:"amount"` //转账数量
	To     string `json:"to"`     //目标地址
}

type CmderBee struct {
	MsgType
	Cmd    BeeCmdType `json:"cmd"` //1 是启动bzz节点，2 是关闭bzz节点，3 是钱包转账
	Url    string     `json:"url"`
	Params []BeeParamArgs
}

func GetDeKey(id string) interface{} {
	return CmdDeKey{
		MsgType: MsgType{
			ID:   id,
			Type: 50,
		},
	}
}

type CmdDeKey struct {
	MsgType
	Key string `json:"key"`
}

type BlockCheck struct {
	MsgType
	Cid string `json:"cid"`
}
