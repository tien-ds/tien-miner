package chia

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/buger/jsonparser"
	"github.com/ds/depaas/protocol"

	"github.com/ds/depaas/utils"

	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func ChiaClient(url string, query []byte) ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logrus.Info("httpPostClient home:", home)
	if !Exists(home+"/.chia/mainnet/config/ssl/full_node/private_full_node.crt") || !Exists(home+"/.chia/mainnet/config/ssl/full_node/private_full_node.key") {
		return nil, errors.New("Node not exist!")
	}
	ce, err := tls.LoadX509KeyPair(home+"/.chia/mainnet/config/ssl/full_node/private_full_node.crt", home+"/.chia/mainnet/config/ssl/full_node/private_full_node.key")
	//logrus.Debug(err)
	tlsConf := &tls.Config{
		//RootCAs: caCertPool,
		Certificates: []tls.Certificate{
			ce,
		},
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConf}
	client := &http.Client{Transport: tr}

	reader := bytes.NewReader(query)

	//logrus.Debug("httpPostClient url:", url, ",query:", string(query))
	request, err := http.NewRequest("POST", url, reader)
	defer request.Body.Close()

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(request) //Do 方法发送请求，返回 HTTP 回复
	if err != nil {
		logrus.Error("httpPostClient err:", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	logrus.Debugf("httpPostClient respBytes %s", string(respBytes))
	return respBytes, err
}

type BlockchainInfo struct {
	protocol.MsgType
	PeerID     string      `json:"peerId"`
	Difficulty int         `json:"difficulty"`
	Height     int         `json:"height"`
	Timestamp  interface{} `json:"timestamp"`
	Balance    int64       `json:"balance"`
	Address    string      `json:"address"`
	State      bool        `json:"state"`
	SyncMode   bool        `json:"sync_mode"`
	Network    string      `json:"network"`
	Space      int64       `json:"space"`
	Size       int64       `json:"size"`
	ChiaID     string      `json:"chiaId"`
}

func ChiaInfo() (BlockchainInfo, error) {
	type ChiaBlockchainState struct {
		BlockchainState struct {
			Difficulty                  int  `json:"difficulty"`
			GenesisChallengeInitialized bool `json:"genesis_challenge_initialized"`
			MempoolSize                 int  `json:"mempool_size"`
			Peak                        struct {
				ChallengeBlockInfoHash string `json:"challenge_block_info_hash"`
				ChallengeVdfOutput     struct {
					Data string `json:"data"`
				} `json:"challenge_vdf_output"`
				Deficit                            int         `json:"deficit"`
				FarmerPuzzleHash                   string      `json:"farmer_puzzle_hash"`
				Fees                               interface{} `json:"fees"`
				FinishedChallengeSlotHashes        interface{} `json:"finished_challenge_slot_hashes"`
				FinishedInfusedChallengeSlotHashes interface{} `json:"finished_infused_challenge_slot_hashes"`
				FinishedRewardSlotHashes           interface{} `json:"finished_reward_slot_hashes"`
				HeaderHash                         string      `json:"header_hash"`
				Height                             int         `json:"height"`
				InfusedChallengeVdfOutput          struct {
					Data string `json:"data"`
				} `json:"infused_challenge_vdf_output"`
				Overflow                   bool        `json:"overflow"`
				PoolPuzzleHash             string      `json:"pool_puzzle_hash"`
				PrevHash                   string      `json:"prev_hash"`
				PrevTransactionBlockHash   interface{} `json:"prev_transaction_block_hash"`
				PrevTransactionBlockHeight int         `json:"prev_transaction_block_height"`
				RequiredIters              string      `json:"required_iters"`
				RewardClaimsIncorporated   interface{} `json:"reward_claims_incorporated"`
				RewardInfusionNewChallenge string      `json:"reward_infusion_new_challenge"`
				SignagePointIndex          int         `json:"signage_point_index"`
				SubEpochSummaryIncluded    interface{} `json:"sub_epoch_summary_included"`
				SubSlotIters               string      `json:"sub_slot_iters"`
				Timestamp                  interface{} `json:"timestamp"`
				TotalIters                 string      `json:"total_iters"`
				Weight                     string      `json:"weight"`
			} `json:"peak"`
			Space        int64 `json:"space"`
			SubSlotIters int   `json:"sub_slot_iters"`
			Sync         struct {
				SyncMode           bool `json:"sync_mode"`
				SyncProgressHeight int  `json:"sync_progress_height"`
				SyncTipHeight      int  `json:"sync_tip_height"`
				Synced             bool `json:"synced"`
			} `json:"sync"`
		} `json:"blockchain_state"`
		Success bool `json:"success"`
	}

	type AllWalletInfo struct {
		Success bool `json:"success"`
		Wallets []struct {
			Data string `json:"data"`
			ID   int    `json:"id"`
			Name string `json:"name"`
			Type int    `json:"type"`
		} `json:"wallets"`
	}

	type WalletAddress struct {
		Address  string `json:"address"`
		Success  bool   `json:"success"`
		WalletID int    `json:"wallet_id"`
	}

	type WalletBalance struct {
		Success       bool `json:"success"`
		WalletBalance struct {
			ConfirmedWalletBalance   int64 `json:"confirmed_wallet_balance"`
			MaxSendAmount            int64 `json:"max_send_amount"`
			PendingChange            int64 `json:"pending_change"`
			SpendableBalance         int64 `json:"spendable_balance"`
			UnconfirmedWalletBalance int64 `json:"unconfirmed_wallet_balance"`
			WalletID                 int64 `json:"wallet_id"`
		} `json:"wallet_balance"`
	}

	var ret BlockchainInfo
	ret.ChiaID = utils.GetMacWifi()

	stateData, err := ChiaClient("https://localhost:8555/get_blockchain_state", []byte("{\"\":\"\"}"))
	if err != nil {
		return ret, err
	}

	var state ChiaBlockchainState
	err = json.Unmarshal(stateData, &state)

	ret.Height = state.BlockchainState.Peak.Height
	ret.Timestamp = state.BlockchainState.Peak.Timestamp
	ret.SyncMode = state.BlockchainState.Sync.Synced
	ret.Difficulty = state.BlockchainState.Difficulty
	ret.State = true
	ret.Space = state.BlockchainState.Space
	ret.Network = "mainnet"

	allWalletData, err := ChiaClient("https://localhost:9256/get_wallets", []byte("{}"))
	if err != nil {
		return ret, err
	}

	var allWalletInfo AllWalletInfo
	err = json.Unmarshal(allWalletData, &allWalletInfo)
	if err != nil || len(allWalletInfo.Wallets) == 0 {
		logrus.Info("err:", err, ",allWalletInfo:", allWalletInfo)
		return ret, err
	}

	data := make(map[string]interface{})
	data["wallet_id"] = allWalletInfo.Wallets[0].ID
	sendDate, _ := json.Marshal(data)
	allWalletData, err = ChiaClient("https://localhost:9256/get_wallet_balance", sendDate)
	if err != nil {
		return ret, err
	}

	var balance WalletBalance
	err = json.Unmarshal(allWalletData, &balance)
	ret.Balance = balance.WalletBalance.ConfirmedWalletBalance

	data["new_address"] = true
	sendDate, _ = json.Marshal(data)
	allWalletData, err = ChiaClient("https://localhost:9256/get_next_address", sendDate)
	if err != nil {
		return ret, err
	}

	var address WalletAddress
	err = json.Unmarshal(allWalletData, &address)
	ret.Address = address.Address

	allWalletData, err = ChiaClient("https://localhost:8560/get_plots", sendDate)
	if err != nil {
		logrus.Info("err:", err, ",allWalletData:", string(allWalletData))
		return ret, err
	}

	totalCount := 0
	totalSize := int64(0)
	jsonparser.ArrayEach(allWalletData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//fmt.Printf("Value: '%s' Type: %s, offset%d", string(value), dataType, offset)
		size, _ := jsonparser.GetInt(value, "file_size")
		ret.Size += size
		totalSize += size
		totalCount = offset
	}, "plots")

	infodata, _ := json.MarshalIndent(ret, "", "  ")
	logrus.Debug("ret:", string(infodata), ",totalSize:", totalSize, ",totalCount:", totalCount)

	return ret, nil
}
