package ip2region

import (
	"fmt"
	"testing"
)

func TestIp(t *testing.T) {
	region, err := New()
	if err != nil {
		panic(err)
	}
	info, _ := region.BinarySearch("114.114.114.114")
	fmt.Println(info.Country, info.Province)
}
