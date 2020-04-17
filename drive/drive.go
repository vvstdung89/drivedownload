package drive

import (
	"encoding/json"
	"errors"
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

type DriveFileInfo struct {
	Md5       string `json:"md5Checksum"`
	Type      string `json:"mimeType"`
	VideoInfo struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Duration string `json:"durationMillis"`
	} `json:"videoMediaMetadata"`
}

func FileInfo(driveID string) *DriveFileInfo {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	request, _ := http.NewRequest("GET", fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=md5Checksum%%2CmimeType%%2CvideoMediaMetadata&key=%s", driveID, "AIzaSyC1eQ1xj69IdTMeii5r7brs3R90eck-m7k"), nil)
	request.Header.Add("x-origin", "https://drive.google.com")
	request.Header.Add("x-referer", "https://drive.google.com")
	request.Header.Add("referer", "https://drive.google.com")
	response, err := netClient.Do(request)

	if err != nil {
		log.Println(err)
		return nil
	}
	body, _ := ioutil.ReadAll(response.Body)
	info := &DriveFileInfo{}
	fmt.Println(string(body))
	if err = json.Unmarshal(body, info); err != nil {
		log.Println(err)
		return nil
	}
	return info
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

	re := regexp.MustCompile("(DRIVE_STREAM=([^;]+))")
	drive_stream := re.Find([]byte(response.Header["Set-Cookie"][0]))

	body, _ := ioutil.ReadAll(response.Body)
	status := getQueryValue(string(body), "status")
	if status == "ok" {
		streamArr := strings.Split(getQueryValue(string(body), "url_encoded_fmt_stream_map"), ",")
		for _, stream := range streamArr {
			res := getQueryValue(stream, "itag")
			_url := getQueryValue(stream, "url")
			//var re = regexp.MustCompile(`(https://[^\.]*)\.([^\/]*)`)
			//_url = re.ReplaceAllString(_url, `$1.drive.google.com`)
			switch res {
			case "18":
				streamInfo["360"], _ = GetFinalRedirectUrl(_url, string(drive_stream))
			case "59":
				streamInfo["480"], _ = GetFinalRedirectUrl(_url, string(drive_stream))
			case "22":
				streamInfo["720"], _ = GetFinalRedirectUrl(_url, string(drive_stream))
			case "37":
				streamInfo["1080"], _ = GetFinalRedirectUrl(_url, string(drive_stream))
			}
			if expireTime == 0 {
				u, _ := url.Parse(_url)
				query := u.Query()
				createdTime, _ = strconv.ParseInt(query["lmt"][0], 10, 64)
				createdTime = int64(math.Floor(float64(createdTime / 1000)))
				expireTime, _ = strconv.ParseInt(query["expire"][0], 10, 64)
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
			log.Println(driveID, err.Error())
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

func CheckDownloadLink(driveID string) bool {
	location := ""
	var netClient = &http.Client{
		Timeout: time.Second * 5,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	url := fmt.Sprintf("https://drive.google.com/uc?id=%s&confirm=Mzub&export=download", driveID)
	response, err := netClient.Get(url)

	if err != nil {
		//if time out still return true
		log.Println(driveID, err.Error())
		return true
	}
	location = response.Header.Get("location")
	if location != "" {
		log.Println(location)
		return true
	}

	//return false only if no location
	return false
}

func GetFinalRedirectUrl(link string, cookie string) (string, error) {
	req, _ := http.NewRequest("HEAD", link, nil)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)

	if err == nil {
		if resp == nil {
			return link, errors.New("Empty response")
		}
		if resp.StatusCode == 302 {
			return GetFinalRedirectUrl(resp.Header.Get("Location"), cookie)
		} else if resp.StatusCode == 200 || resp.StatusCode == 206 {
			return link, nil
		}
	}
	return link, err
}
