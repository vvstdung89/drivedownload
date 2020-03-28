package drive

import (
	"log"
	"testing"
	"time"
)

func TestCheckDownloadLink(*testing.T) {
	xxx := FileInfo("0B6ofTfK8k1_sSWZYLUo4OEpKajA")
	log.Println(xxx)
}

func TestStreamInfo(t *testing.T) {
	time1 := time.Now()
	_ = GetDriveStream("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	log.Println(time.Since(time1).Seconds())
	time2 := time.Now()
	_ = GetDriveStream("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	log.Println(time.Since(time2).Seconds())
	//xxx = GetDriveStream("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	//log.Println(xxx)
}
