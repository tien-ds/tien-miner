package main

import (
	"flag"
	"fmt"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/closer"
	"github.com/ds/depaas/ipds"
	"github.com/ds/depaas/ipds/web"
	"github.com/ds/depaas/persistence"
	_ "github.com/ds/depaas/persistence/orbitdb"
	"github.com/ds/depaas/register"
	"github.com/ds/depaas/rest"
	"github.com/ds/depaas/service"
	"github.com/ds/depaas/utils"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)

var (
	ip         = flag.String("ip", "", "set gw ip")
	serverAddr = flag.String("listen", ":8099", "listen server addr")
	boot       = flag.String("boot", "", "start boot peer")
)

func init() {
	rest.InitRest()
	register.Init()

	os.Setenv("repo", "gw")
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().SetOnline("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("content-type", "text/plain")
	if websocket.IsWebSocketUpgrade(r) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			logrus.Error(err)
			return
		}
		service.NewMsg(conn).ReadMessage()
	}
}

func InitService() {
	logrus.Info("init client")
	ipds.IPfsInit()

	logrus.Info("init database")
	persistence.InitDataBase(nil, nil)

	logrus.Info("init database success")
	service.SubscribesInit()

	logrus.Info("init database FileServer")
	web.StartFileServer()
}

func main() {
	utils.SetLog(true)

	flag.Parse()

	//set gw ip
	if *ip == "" {
		fmt.Println("ip not set")
		flag.Usage()
		os.Exit(0)
	} else {
		register.SetAppIp(*ip)
	}

	InitService()

	//connect boot
	if *boot != "" {
		ConnectBoot(*boot)
	}

	go utils.Shutdown(func() {
		closer.CloseAll()
	})
	api.PackApi()
	logrus.Infof("listen %s", *serverAddr)
	http.HandleFunc("/p2p/", wsHandler)
	http.HandleFunc("/api/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		api.ApiHttp(writer, request, nil)
	})
	//http.Handle("/", http.FileServer(web.Asset))
	log.Fatal(http.ListenAndServe(*serverAddr, nil))
}
