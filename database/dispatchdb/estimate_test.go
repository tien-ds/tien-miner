package dispatchdb

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	for i := 1; i <= 5; i++ {
		fmt.Println(Store(time.Now().String(), "bb", uint64(i*10)))
	}
	fmt.Println(EstimateRepo("bb"))
}
