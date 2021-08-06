package env

import (
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	duration, err := time.ParseDuration("600s")
	fmt.Println(duration, err)
}
