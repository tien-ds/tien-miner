package rest

import (
	"fmt"
	"gitee.com/fast_api/api"
	"github.com/ds/depaas/service"
)

func test() {
	api.POST(func(a struct {
		Peer string `json:"peer,omitempty"`
		Cid  string `json:"cid"`
	}) {
		service.NewPinServer(a.Peer).PinAdd(a.Cid, func(id string, result interface{}) {
			fmt.Println(result)
		})
	}, "/api/pinAdd")

	api.POST(func(a struct {
		Peer string `json:"peer,omitempty"`
		Cid  string `json:"cid"`
	}) {
		service.NewPinServer(a.Peer).PinRm(a.Cid, func(id string, result interface{}) {
			fmt.Println(result)
		})
	}, "/api/pinRm")

	api.POST(func(a struct {
		Peer string `json:"peer,omitempty"`
		Cid  string `json:"cid"`
		File string `json:"file"`
	}) {
		service.NewPinServer(a.Peer).PinGet(a.Cid, a.File, func(id string, result interface{}) {
			fmt.Println(result)
		})
	}, "/api/get")
}
