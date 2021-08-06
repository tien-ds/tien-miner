package node

//
//import (
//	"github.com/ipfs/go-ipfs/addrassets"
//	"github.com/sirupsen/logrus"
//	"net/http"
//	"strings"
//)
//
//func StartApi(apiUrl string, msg chan int) {
//	logrus.Debug("addressBind apiAddr:", apiUrl)
//	http.Handle("/api/v0/set/passwd", &addressHandler{apiUrl: apiUrl, msg: msg})
//	http.Handle("/api/v0/set/address", &addressHandler{apiUrl: apiUrl, msg: msg})
//	http.Handle("/api/v0/set/query", &addressHandler{apiUrl: apiUrl, msg: msg})
//	http.Handle("/api/v0/peerId", &addressHandler{apiUrl: apiUrl, msg: msg})
//	http.Handle("/", http.FileServer(addrassets.AssetFile()))
//	http.ListenAndServe(":12345", nil)
//}
//
//type addressHandler struct {
//	apiUrl string
//	msg    chan int
//}
//
//func (ah *addressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	url := ah.apiUrl + r.URL.String()
//	res, _ := HttpPostClient(url, nil)
//
//	w.Header().Set("Access-Control-Allow-Origin", "*")
//
//	w.Write(res)
//	logrus.Debug("ServeHTTP r.URL.String():", r.URL.String())
//	if strings.HasPrefix(r.URL.String(), "/api/v0/set/address") {
//		logrus.Debug("ServeHTTP msg r.URL.String():", r.URL.String())
//		ah.msg <- 0
//	}
//}
