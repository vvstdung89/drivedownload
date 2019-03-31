package drive

import (
	"errors"
	"github.com/vvstdung89/goutils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func DownloadStream(driveID string, quality string, dst string) error {
	streamInfo := GetDriveStream(driveID, "")
	streamLink := streamInfo.Streams[quality]
	cookie := streamInfo.Cookie

	client := &http.Client{
		Jar: goutils.NewJar(),
	}
	u, err := url.Parse(streamLink)
	cookies := []*http.Cookie{}
	cookies = append(cookies, &http.Cookie{Name: "DRIVE_STREAM", Value: strings.Split(cookie, "=")[1]})
	client.Jar.SetCookies(u, cookies)
	req, _ := http.NewRequest("GET", u.String(), nil)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	chunk := 2000000
	parallel := 16

	if size == 0 {
		return errors.New("Cannot get file size")
	} else {
		return goutils.ChunkDownload(dst, streamLink, strings.Split(cookie, "=")[1], size, chunk, parallel)
	}
}
