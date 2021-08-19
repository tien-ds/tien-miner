package ipds

import (
	"testing"

	"gitee.com/fast_api/api"
)

func TestIpfs(t *testing.T) {

	api.StartService(":8066")
}
