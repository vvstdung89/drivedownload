package error

import (
	"log"
	"testing"
)

var (
	HOWDP = ECode{10, "howdo"}
)

func TestE(t *testing.T) {
	log.Println(AdvError{E(HOWDP, "BarError"), "123"})
}
