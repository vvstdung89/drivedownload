package goutils

import (
	"fmt"
	"github.com/vvstdung89/goutils/drive"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestChunkDownload(*testing.T) {
	streamInfo := drive.GetDriveStream("1_wdQ5mnVZxcO4OPBDXWJOPUwjzl25DUD", "")
	//fmt.Println(streamInfo.Streams["720"], streamInfo.Cookie)
	dst := "/tmp/filetmp"
	streamLink := streamInfo.Streams["720"]
	cookie := streamInfo.Cookie

	client := &http.Client{
		Jar: NewJar(),
	}
	u, err := url.Parse(streamLink)
	cookies := []*http.Cookie{}
	cookies = append(cookies, &http.Cookie{Name: "DRIVE_STREAM", Value: strings.Split(cookie, "=")[1]})
	client.Jar.SetCookies(u, cookies)
	fmt.Println(strings.Split(cookie, "=")[1])
	req, _ := http.NewRequest("GET", u.String(), nil)

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	chunk := 2000000
	parallel := 16

	if size == 0 {
		//log.Println(url, cookie)
		log.Println(resp.Status)
		log.Println(resp.Header)
	} else {
		if err := ChunkDownload(dst, streamLink, strings.Split(cookie, "=")[1], size, chunk, parallel); err != nil {
			log.Println(err)
		} else {
			log.Println("Download success")
		}
	}

}
