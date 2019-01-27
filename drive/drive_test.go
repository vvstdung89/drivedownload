package drive

import (
	"github.com/vvstdung89/goutils/drive"
	"log"
	"testing"
)

func TestCheckDownloadLink(*testing.T) {
	xxx := drive.CheckDownloadLink("ddd")
	log.Println(xxx)
}
