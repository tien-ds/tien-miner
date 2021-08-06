package ipds

import (
	"testing"

	"gitee.com/fast_api/api"
)

func TestIpfs(t *testing.T) {
	GwInit()
	api.StartService(":8066")
}
