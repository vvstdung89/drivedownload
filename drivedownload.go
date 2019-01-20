package drivedownload


import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DriveStreamInfo struct {
	Cookie      string
	Streams     map[string]string
	CreatedTime int64
	ExpireTime  int64
}

type DriveDownInfo struct {
	Link       string
	ExpireTime int64
}

func getQueryValue(queryString, key string) string {
	query, err := url.ParseQuery(queryString)
	if err != nil {
		return ""
	}
	if query[key] != nil {
		return query[key][0]
	}
	return ""
}

func StreamInfo(driveID string, accessToken string) DriveStreamInfo {

	//get from cache
	var driveStreamInfo = DriveStreamInfo{
		ExpireTime: time.Now().Add(15 * time.Minute).Unix(),
	}

	var streamInfo = make(map[string]string)
	var createdTime, expireTime int64
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	response, err := netClient.Get(fmt.Sprintf("https://drive.google.com/e/get_video_info?docid=%s&access_token=%s", driveID, accessToken))
	if err != nil {
		return driveStreamInfo
	}

	re := regexp.MustCompile("(DRIVE_STREAM=([^;]+);)")
	drive_stream := re.Find([]byte(response.Header["Set-Cookie"][0]))

	body, _ := ioutil.ReadAll(response.Body)
	status := getQueryValue(string(body), "status")
	if status == "ok" {
		streamArr := strings.Split(getQueryValue(string(body), "url_encoded_fmt_stream_map"), ",")
		for _, stream := range streamArr {
			res := getQueryValue(stream, "itag")
			_url := getQueryValue(stream, "url")
			var re = regexp.MustCompile(`(https://[^\.]*)\.([^\/]*)`)
			_url = re.ReplaceAllString(_url, `$1.gvt1.com`)
			switch res {
			case "18":
				streamInfo["360"] = _url
				u, _ := url.Parse(_url)
				query := u.Query()
				createdTime, _ = strconv.ParseInt(query["lmt"][0], 10, 64)
				createdTime = int64(math.Floor(float64(createdTime / 1000)))
				expireTime, _ = strconv.ParseInt(query["expire"][0], 10, 64)
			case "59":
				streamInfo["480"] = _url
			case "22":
				streamInfo["720"] = _url
			case "37":
				streamInfo["1080"] = _url
			}
		}
		driveStreamInfo = DriveStreamInfo{Cookie: string(drive_stream), Streams: streamInfo, CreatedTime: createdTime, ExpireTime: expireTime}
	}
	return driveStreamInfo
}

func DownloadInfo(driveID string, accessToken string) DriveDownInfo {

	var driveDownInfo = DriveDownInfo{
		ExpireTime: time.Now().Add(15 * time.Minute).Unix(),
	}

	var netClient = &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	location := ""
	if accessToken == "" {
		url := fmt.Sprintf("https://drive.google.com/uc?id=%s&confirm=Mzub&export=download", driveID)
		response, err := netClient.Get(url)
		if err != nil {
			log.Println(driveID, accessToken, err.Error())
			return driveDownInfo
		}
		location = response.Header.Get("location")
	} else {
		url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media&access_token=%s", driveID, accessToken)
		response, err := netClient.Get(url)
		if err != nil {
			log.Println(driveID, accessToken, err.Error())
			return driveDownInfo
		}
		location = response.Header.Get("location")
	}

	driveDownInfo.Link = location
	if location != "" {
		driveDownInfo.ExpireTime = time.Now().Add(time.Hour * 2).Unix()
	} else {
		driveDownInfo.ExpireTime = time.Now().Add(time.Minute * 10).Unix()
	}

	return driveDownInfo
}
