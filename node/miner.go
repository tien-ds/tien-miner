package node

import (
	"bytes"
	"context"
	"debug/elf"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/ds/depaas/database/config"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/ipds/muldisk"
	service2 "github.com/ds/depaas/ipds/service"
	"github.com/ds/depaas/node/bee"
	"github.com/ds/depaas/node/chia"
	"github.com/ds/depaas/node/env"
	nutils "github.com/ds/depaas/node/utils"
	"github.com/ds/depaas/protocol"
	"github.com/ds/depaas/utils"
	"github.com/gorilla/websocket"
	"github.com/ipfs/go-cid"
	cmds "github.com/ipfs/go-ipfs-cmds"
	oldcmds "github.com/ipfs/go-ipfs/commands"
	"github.com/multiformats/go-multiaddr"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
	"github.com/tien-ds/contract-miner/miner"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type NodeContext struct {
	Context       context.Context
	msg           chan int
	apiAddr       string
	p2p           string
	peerId        string
	downloadSpeed float64
	uploadSpeed   float64
	msgId         string
	conn          *websocket.Conn
	version       int
	runVersion    int
	mnemonic      string
	mi            *miner.MinerEx
}

func (ws *NodeContext) Write(p []byte) (int, error) {
	info := make(map[string]interface{})
	info["peerId"] = ws.peerId
	info["type"] = 9
	info["text"] = string(p)
	data, _ := json.Marshal(info)
	err := ws.WriteMessage(makeAesMsg(data))
	logrus.Debug("Write p:", string(p), ",data:", string(data))
	return len(p), err
}

func (ws *NodeContext) echoError(err error) error {
	info := make(map[string]interface{})
	info["peerId"] = ws.peerId
	info["type"] = 9
	info["text"] = ""

	if err != nil {
		logrus.Debug("run err:", err)
		info["text"] = "error:" + err.Error()
	}
	return ws.SendEncryptMessage(info)
}

func (ws *NodeContext) runShellCmd(params []string) error {

	logrus.Debug("run params:", params)

	cmd := exec.Command(params[0], params[1:]...)
	//cmd.Stdin = strings.NewReader("abcdefg")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	cmd.Stdout = ws
	err := cmd.Run()

	return ws.echoError(err)
}

func (ws *NodeContext) runCmds(params [][]string) error {
	logrus.Debug("runCmds params:", params)

	for i := 0; i < len(params); i++ {
		cmd := exec.Command(params[i][0], params[i][1:]...)
		//cmd.Stdin = strings.NewReader("abcdefg")
		//var out bytes.Buffer
		//cmd.Stdout = &out
		//cmd.Stdout = ws
		err := cmd.Run()
		if err != nil {
			return ws.echoError(err)
		}
	}

	return ws.echoError(nil)
}

func httpPost(url string, info []byte) (*http.Response, int64, error) {
	client := new(http.Client)
	reader := bytes.NewReader(info)

	start := time.Now()
	request, _ := http.NewRequest("POST", url, reader)
	defer request.Body.Close()

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(request) //Do 方法发送请求，返回 HTTP 回复

	elapsed := time.Since(start)
	var time = elapsed.Nanoseconds()
	logrus.Debug("httpPost elapsed:", elapsed, ",time:", time)
	return resp, time, err
}

func makeAesMsg(msg []byte) []byte {
	if msg == nil {
		return nil
	}
	key := utils.AesPasswd()
	encrypted := utils.AesEncryptCBC(msg, key)

	logrus.Debug("MAKE_AES_MSG origin:", string(msg))
	d := protocol.AesType{
		MsgType: protocol.MsgType{
			Type: protocol.AES_ENCRYPT,
		},
		Msg: base64.StdEncoding.EncodeToString(encrypted),
	}
	byts, _ := json.Marshal(d)
	return byts
}

func decodeAesMsg(msg []byte) ([]byte, error) {
	str, err := jsonparser.GetString(msg, "msg")
	if err != nil {
		return nil, err
	}

	deData, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		logrus.Debug("decodeAesMsg base64 str:", str, ",err:", err)
		return nil, err
	}

	key := utils.AesPasswd()
	logrus.Debug("decodeAesMsg AesDecryptCBC str:", str)
	decrypted := utils.AesDecryptCBC(deData, key)
	logrus.Debug("decodeAesMsg msg:", string(msg))
	return decrypted, nil
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func diskTotalSpace() (float64, []disk.PartitionStat) {
	diskInfo := muldisk.GetTotalDisk()
	return float64(diskInfo.Uint64()) / 1_000_000_000, nil
}

func (ws *NodeContext) machineInfo() protocol.InfoType {
	v, _ := mem.VirtualMemory()
	infos, _ := cpu.Info()
	_, _, _, totalIn, totalOut, repoSize := statsBwAndRepoSize()
	totalSpace, _ := diskTotalSpace()
	return protocol.InfoType{
		MsgType: protocol.MsgType{
			ID:   ws.msgId,
			Type: protocol.MINFO,
		},
		PeerId:      ws.peerId,
		CPUCount:    len(infos),
		DiskTotal:   totalSpace,
		IpOnLine:    false,
		RAM:         int64(v.Total),
		MachineType: int(nutils.GetDsType()),
		TotalIn:     totalIn,
		TotalOut:    totalOut,
		RepoSize:    int64(repoSize),
		Addr:        config.GetBindAddr(),
		//Ip:          "",
	}
}

func (ws *NodeContext) PowerMiner() {
	//ip, _ := config.GetConfigKeyString(commands.IpConfigKey)
	_, _, _, totalIn, totalOut, repoSize := statsBwAndRepoSize()
	infos, _ := cpu.Info()

	totalSpace, _ := diskTotalSpace()
	v, _ := mem.VirtualMemory()
	info := miner.InfoType{
		CPUCount:    len(infos),
		DiskTotal:   totalSpace,
		DownSpeed:   ws.downloadSpeed,
		UpSpeed:     ws.uploadSpeed,
		RAM:         int64(v.Total),
		MachineType: int(nutils.GetDsType()),
		TotalIn:     totalIn,
		TotalOut:    totalOut,
		Time:        0,
		RepoSize:    int64(repoSize),
		Addr:        config.GetBindAddr(),
		//Ip:          ip,
	}

	bytes, _ := json.Marshal(info)
	ws.mi.PowerStorage(string(bytes))
}

func (ws *NodeContext) chiaInfo(id string) []byte {
	info, _ := chia.ChiaInfo()
	info.Type = 6
	info.ID = id
	info.PeerID = ws.peerId

	data, _ := json.Marshal(info)
	return data
}

func statsBwAndRepoSize() (error, float64, float64, int64, int64, uint64) {
	bw := service2.StatsBw()
	size := service2.RepoSize()
	return nil, bw.RateIn, bw.RateOut, bw.TotalIn, bw.TotalOut, size.RepoSize
}

func (ws *NodeContext) processCMDMsg(msg protocol.SendEntry, id string) error {
	var err error
	var res string
	if len(msg.Params) == 0 {
		return errors.New("param == 0")
	}
	param := msg.Params[0]
	if param.Key != "arg" {
		return errors.New("param key error")
	}
	cidValue := param.Value
	cid, err := cid.Decode(cidValue)
	if err != nil {
		logrus.Error(err)
		return err
	}
	switch msg.Cmd {
	case protocol.PINADD:
		res, err = service2.PinAdd(cid)
		if err == nil {
			config.StoreBlock(cid)
		}
	case protocol.PINRM:
		res, err = service2.PinRm(cid)
	case protocol.CAT, protocol.GET:
		file := msg.Params[1]
		if file.Key == "file" && file.Value != "" {
			err = service2.GetFile(cid, file.Value)
		}
	default:
		logrus.Error("processCMDMsg cmd err msg.Cmd:", msg.Cmd)
		return err
	}

	resultType := protocol.ResultType{
		MsgType: protocol.MsgType{
			ID:   id,
			Type: protocol.CMD_SYSTEM_RESP,
		},
		RandID: msg.RandId,
	}
	if err != nil {
		resultType.Error = err.Error()
	}
	if res != "" {
		resultType.Result = res
	}
	return ws.SendEncryptMessage(resultType)
}

func (ws *NodeContext) WriteMessage(data []byte) error {
	if ws.conn != nil && data != nil {
		logrus.Debugf("SEND MSG %s", string(data))
		return ws.conn.WriteMessage(websocket.TextMessage, data)
	}
	return errors.New("Not connected")
}

func (ws *NodeContext) runPeriod() error {

	err := ws.SendEncryptMessage(ws.machineInfo())
	if err != nil {
		logrus.Debug("SendEncryptMessage ", err)
		return err
	}

	if nutils.GetDsType() == protocol.MINER_HOME {
		id := ws.msgId
		ws.SendEncryptMessage(chia.ChiaInfoMsg(id, ws.peerId))
		ws.SendEncryptMessage(bee.BeeInfo(id, ws.peerId))
	}
	return err
}

func (ws *NodeContext) SendEncryptMessage(f interface{}) error {
	protocol.SetMsgTypeID(f, ws.msgId)
	data, e := json.Marshal(f)
	if e != nil {
		return e
	}
	logrus.Debug("SendMessage ", string(data))
	return ws.WriteMessage(makeAesMsg(data))
}

//func (ws *NodeContext) SendMessage(f interface{}) error {
//	data, e := json.Marshal(f)
//	if e != nil {
//		logrus.Error(e)
//		return e
//	}
//	logrus.Debug("SendMessage ", string(data))
//	return ws.WriteMessage(data)
//}

func (ws *NodeContext) CloseConn() {
	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
	}
}

func getMnemonic() (string, error) {
	data, err := chia.ChiaClient("https://localhost:9256/get_public_keys", []byte("{}"))
	if err != nil {
		return "", err
	}
	fingerprint, err := jsonparser.GetInt(data, "public_key_fingerprints", "[0]")
	logrus.Debug("getMnemonic fingerprint:", fingerprint, ",err:", err)
	if err != nil {
		return "", err
	}
	info := make(map[string]int64)
	info["fingerprint"] = fingerprint
	senddate, _ := json.Marshal(info)
	data, err = chia.ChiaClient("https://localhost:9256/get_private_key", senddate)
	logrus.Debug("getMnemonic data:", string(data), ",err:", err)
	if err != nil {
		return "", err
	}

	ret, err := jsonparser.GetString(data, "private_key", "seed")
	logrus.Debug("getMnemonic ret:", ret, ",err:", err)
	return ret, err
}

func httpDownload(url string, name string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	os.Remove(name)
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	io.Copy(f, res.Body)
	return nil
}

func (ws *NodeContext) updateEcho(id string, code int) error {
	var ret protocol.CmderUpdateRet
	ret.ID = id
	ret.Type = 23
	ret.Code = code
	return ws.SendEncryptMessage(ret)
}

func (ws *NodeContext) update(data []byte) error {
	var cmd protocol.CmderUpdate

	err := json.Unmarshal(data, &cmd)
	if err != nil {
		return nil
	}

	if ws.runVersion > 0 {
		return ws.updateEcho(cmd.ID, 2)
	} else if ws.version >= cmd.Version {
		return ws.updateEcho(cmd.ID, 0)
	}

	const dspath string = "/root/.ds/ds"
	if runtime.GOARCH == "arm64" {
		err = httpDownload(cmd.Arm64, dspath)
		if err == nil {
			exe, err := elf.Open(dspath)
			if err != nil || elf.EM_AARCH64 != exe.Machine {
				os.Remove(dspath)
				return ws.updateEcho(cmd.ID, 4)
			}
		}
	} else {
		err = httpDownload(cmd.Amd64, dspath)
		if err == nil {
			exe, err := elf.Open(dspath)
			if err != nil || elf.EM_X86_64 != exe.Machine {
				return ws.updateEcho(cmd.ID, 4)
			}
		}
	}

	logrus.Debug("update cmd:", cmd, ",err:", err)
	if err != nil {
		return ws.updateEcho(cmd.ID, 3)
	}

	ws.updateEcho(cmd.ID, 1)
	_, err = nutils.HttpPostClient("http://127.0.0.1:9999/cmd", []byte("[[\"pm2\",\"stop\",\"ds\"],[\"mv\",\"-f\",\"/root/.ds/ds\",\"/bin/ds\"],[\"chmod\",\"+x\",\"/bin/ds\"],[\"pm2\",\"start\",\"ds\"]]"))
	if err != nil {
		return ws.updateEcho(cmd.ID, 3)
	}

	return ws.updateEcho(cmd.ID, 1)
}

func (ws *NodeContext) DeeTransfer(url string, params []protocol.BeeParamArgs) (string, error) {
	ret := make(map[string]interface{})
	for _, v := range params {
		//logrus.Debug(index, "\t",value)
		amount := big.NewInt(0)
		amount.SetString(v.Amount, 10)

		hash, err := ws.mi.Transfer(v.To, amount)

		logrus.Debug("DeeTransfer hash:", hash, ",err:", err)
		if err != nil {
			return "", err
		}
		ret[v.To] = hash
	}

	retData, _ := json.Marshal(ret)
	return string(retData), nil
}

func (ws *NodeContext) run() error {

	//step 1
	//connect to gw
	c, _, err := websocket.DefaultDialer.DialContext(ws.Context, ws.p2p, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("connecting to %s %s", ws.p2p, err))
	} else {
		logrus.Infof("connecting to %s success", ws.p2p)
	}
	ws.conn = c
	msgId := make(chan string)

	//step 3
	//Get msg Id
	go func() {
		var hello protocol.MsgType
		_, message, err := c.ReadMessage()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Info("ReadMessage message:", string(message))
		err = json.Unmarshal(message, &hello)
		if err != nil {
			logrus.Error(err)
		}

		msgId <- hello.ID

	}()

	//step 2
	//echo hello
	hello := protocol.HelloType{
		MsgType: protocol.MsgType{ID: "122", Type: protocol.ID_CODE},
		Msg:     "connect",
	}
	ws.SendEncryptMessage(hello)

	//wait for peerId
	ws.msgId = <-msgId

	defer ws.CloseConn()
	logrus.Info("exchange ok")

	//step 4
	//send self miner bind
	minMsg := protocol.Miner{
		MsgType: protocol.MsgType{
			ID:   ws.msgId,
			Type: protocol.MINER_PEER,
		},
		PeerID: ws.peerId,
		Addr:   config.GetBindAddr(),
	}
	ws.SendEncryptMessage(minMsg)

	if ipds.GetNode() == nil {
		ws.SendEncryptMessage(protocol.Message{
			MsgType: protocol.MsgType{
				ID:   ws.msgId,
				Type: protocol.MESSAGE,
			},
			SType: "ERROR",
			State: 0,
			MSG:   "wait for mount disk",
		})
		return errors.New("node nil maybe wait for mount")
	} else {
		ws.SendEncryptMessage(protocol.Message{
			MsgType: protocol.MsgType{
				ID:   ws.msgId,
				Type: protocol.MESSAGE,
			},
			SType: "INFO",
			State: 1,
		})
	}

	//loop read message
	done := make(chan struct{})
	go ws.loopMsg(c, done)

	//miner startup run
	ws.runOnce()

	ws.runPeriod()

	ws.PowerMiner()

	duration, err := time.ParseDuration(env.GetEnv("INTERVAL_TIME"))
	if err != nil {
		logrus.Error(err)
		os.Exit(0)
	}
	logrus.Infof("INTERVAL_TIME %s", duration)
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ws.Context.Done():
			return nil
		case <-done:
			return errors.New("close by remote")
		case <-ws.msg:
			logrus.Debug("webSocketClient run msg:")
			minMsg.Addr = config.GetBindAddr()
			minMsg.MinerAddr = config.GetChainAddress()
			minMsg.Type = 16
			err = ws.SendEncryptMessage(minMsg)
			if err != nil {
				return errors.New(fmt.Sprintf("write tick msg err:", err))
			}
		case <-ticker.C:
			logrus.Debug("webSocketClient run time:")
			ws.PowerMiner()
			ws.runPeriod()

			/*
				case <-interrupt:
					logrus.Debug("webSocketClient interrupt")

					// Cleanly close the connection by sending a close message and then
					// waiting (with timeout) for the server to close the connection.
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						logrus.Debug("webSocketClient write close:", err)
						return
					}
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					return*/
		}
	}
}

// StartLoop wait for loop in reconnect
func (ws *NodeContext) StartLoop() error {

	time.Sleep(1 * time.Second)
	for {
		err := ws.run()
		if err == nil {
			break
		} else {
			logrus.Error(err)
		}
		time.Sleep(5 * time.Second)
	}

	return errors.New("ws client quit!")
}

func (ws *NodeContext) loopMsg(c *websocket.Conn, done chan struct{}) {
	var data protocol.MsgType
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logrus.Error("webSocketRun read:", err)
			break
		}
		err = json.Unmarshal(message, &data)
		if err != nil {
			logrus.Error("webSocketRun Unmarshal:", err, ",message:", message, ",str:", string(message))
			continue
		}

		if data.Type == protocol.AES_ENCRYPT {
			message, err = decodeAesMsg(message)
			if err != nil {
				logrus.Error("webSocketRun decodeAesMsg err:", err)
				ws.echoError(errors.New("decodeAesMsg error"))
				continue
			}

			err = json.Unmarshal(message, &data)
			if err != nil {
				logrus.Error("webSocketRun decodeAesMsg Unmarshal err:", err)
				continue
			}
		}

		logrus.Debug("ReadMessage:", string(message))
		ws.msgId = data.ID
		switch data.Type {
		case protocol.HELLO_CODE:
			var hello protocol.HelloType
			err = json.Unmarshal(message, &hello)
			if hello.Msg != "" {
				bys, err := base64.StdEncoding.DecodeString(hello.Msg)
				if err != nil {
					logrus.Error(err)
					os.Exit(0)
				}
				multiAddr, err := multiaddr.NewMultiaddrBytes(bys)
				if err != nil {
					logrus.Error(err)
					os.Exit(0)
				}
				err = service2.ConnectPeer([]string{multiAddr.String()})
				if err != nil {
					logrus.Error(err)
					os.Exit(0)
				}
				logrus.Infof("connect to %s", multiAddr.String())
			} else {
				os.Exit(0)
			}
		case protocol.CMD_CODE:
			var cmd protocol.Cmder
			err = json.Unmarshal(message, &cmd)
			ws.msgId = data.ID
			ws.runShellCmd(cmd.Params)

		case protocol.MNEMONIC_CODE:
			info := make(map[string]interface{})
			info["peerId"] = data.ID
			info["type"] = 21
			if len(ws.mnemonic) == 0 {
				mne, _ := getMnemonic()
				info["mnemonic"] = mne
				ws.mnemonic = mne
			} else {
				info["mnemonic"] = ws.mnemonic
			}
			err = ws.SendEncryptMessage(info)
			if err != nil {
				break
			}

		case protocol.SELF_UPDATE_CODE:
			if nil != ws.update(message) {
				break
			}

		case protocol.CHIA_INFO_CODE:
			err = ws.SendEncryptMessage(chia.ChiaInfoMsg(data.ID, ws.peerId))
			if err != nil {
				break
			}
			err = ws.SendEncryptMessage(bee.BeeInfo(data.ID, ws.peerId))
			if err != nil {
				logrus.Debug("runPeriod BeeInfo err:", err)
			}

		case protocol.BEE_CMD_CODE:
			var cmd protocol.CmderBee
			err = json.Unmarshal(message, &cmd)
			ws.msgId = cmd.ID
			if cmd.Cmd == protocol.BEE_START {
				ws.runCmds([][]string{{"docker", "start", "bzz2"}})
			} else if cmd.Cmd == protocol.BEE_STOP {
				ws.runCmds([][]string{{"docker", "stop", "bzz2"}})
			} else if cmd.Cmd == protocol.BEE_TRANSFER || cmd.Cmd == protocol.DE_TRANSFER {
				var err error
				var result string
				if cmd.Cmd == protocol.BEE_TRANSFER {
					result, err = bee.BeeTransfer(cmd.Url, cmd.Params)
				} else {
					result, err = ws.DeeTransfer(cmd.Url, cmd.Params)
				}
				cmdResult := protocol.CmderResult{
					MsgType: protocol.MsgType{ID: cmd.ID},
				}
				if err == nil {
					cmdResult.Type = protocol.CMD_RESP_SUCCESS
					cmdResult.Text = result
				} else {
					cmdResult.Type = protocol.CMD_RESP_FAIL
					cmdResult.Text = err.Error()
				}
				err = ws.SendEncryptMessage(cmdResult)
				if err != nil {
					break
				}
			}

		case 50:
			var key protocol.CmderPrikey
			key.ID = data.ID
			key.Type = 51
			key.Key = config.GetChainPrivateKey()
			key.Address = config.GetChainAddress()
			err = ws.SendEncryptMessage(key)
			if err != nil {
				logrus.Debug("runPeriod BeeInfo err:", err)
			}

		case protocol.CMD_SYSTEM:
			var enterData protocol.SendEntry
			err = json.Unmarshal(message, &enterData)
			if err != nil {
				logrus.Debug("webSocketRun enterData Unmarshal:", err, ",message:", message, ",str:", string(message))
				continue
			}
			ws.processCMDMsg(enterData, data.ID)
		case protocol.MINER_DISK_CODE:

		}
	}
}

func (ws *NodeContext) runOnce() {

	if nutils.GetDsType() == protocol.MINER_HOME {
		w := protocol.Wifi{
			MsgType: protocol.MsgType{Type: protocol.WIFI_KEY_PAIR},
			PeerID:  ws.peerId,
		}
		mac := utils.GetMacWifi()
		if len(mac) >= 4 {
			wifiName := "DS" + mac[4:]
			w.Wifi = wifiName
			logrus.Debug("webSocketRun wifiName:", wifiName)
		}
		ws.SendEncryptMessage(w)

		//wifi bind
		//bee.BeeInfo("123", ws.peerId)
	}

}

func NewWsClient(ctx context.Context, p2p string, peerId string, msg chan int, mi *miner.MinerEx) *NodeContext {
	//TODO chia Mnemonic
	//mne, _ := getMnemonic()
	return &NodeContext{
		Context:       ctx,
		p2p:           p2p,
		peerId:        peerId,
		msg:           msg,
		downloadSpeed: 0,
		uploadSpeed:   0,
		version:       1008,
		runVersion:    0,
		//mnemonic:      mne,
		mi: mi,
	}
}

// StartMinerWithNode this method will wait for loop
func StartMinerWithNode(ctx context.Context) error {
	peerId := ipds.GetPeerID()
	logrus.Infof("peerID %s", peerId)

	p2p := env.GetEnv("P2P")
	logrus.Debugf("P2P ADDR %s ", p2p)

	msg := make(chan int, 10)

	//go func() {
	//	StartApi("http://127.0.0.1", msg)
	//}()

	//init contract miner
	miner.InitMiner(env.GetEnv("MINER_RPC"))
	mi := miner.NewMinerEx(env.GetEnv("MINER_CONTRACT"), config.GetChainPrivateKey())

	//set contract addr map
	err := mi.SetMap(peerId)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Debugf("Address: %s", config.GetChainAddress())

	ws := NewWsClient(ctx, p2p, peerId, msg, mi)
	//set node
	SetNodeContext(ws)

	return ws.StartLoop()
}

func StartMiner(req *cmds.Request, ctx *oldcmds.Context) error {
	node, err := ctx.ConstructNode()
	if err != nil {
		logrus.Error(err)
		return err
	}
	ipds.SetNode(node)
	return StartMinerWithNode(req.Context)
}
