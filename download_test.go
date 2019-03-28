package goutils

import (
	"github.com/vvstdung89/goutils/drive"
	"log"
	"testing"
)

func TestChunkDownload(*testing.T) {

	streamInfo := drive.GetDriveStream("0B0EM1NfwGeVfMEtCVkRZcWh2QnM", "")
	//fmt.Println(streamInfo.Streams["720"], streamInfo.Cookie)
	dst := "/tmp/filetmp"
	url := streamInfo.Streams["720"]
	cookie := streamInfo.Cookie
	size := 144916608
	chunk := 2000000
	parallel := 8

	if err := ChunkDownload(dst, url, cookie, size, chunk, parallel); err != nil {
		log.Println(err)
	} else {
		log.Println("Download success")
	}

}
