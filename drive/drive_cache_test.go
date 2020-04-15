package drive

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestGetDriveRes(t *testing.T) {
	time1 := time.Now()
	x := GetDriveRes("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	fmt.Println(x)
	log.Println(time.Since(time1).Seconds())
	time2 := time.Now()
	_ = GetDriveRes("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	log.Println(time.Since(time2).Seconds())
	//xxx = GetDriveStream("0B6ofTfK8k1_sSWZYLUo4OEpKajA", "")
	//log.Println(xxx)

}
