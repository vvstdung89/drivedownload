package goutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var GAPIUrl = "https://www.googleapis.com"

type RefreshTokenResult struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

type UploadResult struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UploadMetaData struct {
	Name        string   `json:"name"`
	Parents     []string `json:"parents,omitempty"`
	TeamDriveId string   `json:"teamDriveId,omitempty"`
}

func RefreshToken(refresh_token string, client_id string, secret string) (*RefreshTokenResult, error) {
	var actionURL = GAPIUrl + "/oauth2/v4/token"

	client := &http.Client{}

	params := url.Values{}
	params.Add("client_secret", secret)
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", refresh_token)
	params.Add("client_id", client_id)

	var resp *http.Response
	var err error
	var content []byte

	if resp, err = client.PostForm(actionURL, params); err != nil {
		return nil, err
	}

	if content, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Println(string(content))
		return nil, errors.New("Status code with " + strconv.Itoa(resp.StatusCode))
	}

	var res = &RefreshTokenResult{}

	if err = json.Unmarshal(content, res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetTeamDrive(accesstoken string) ([]string, error) {
	var actionURL = GAPIUrl + "/drive/v3/teamdrives"
	client := &http.Client{}
	var content []byte
	req, err := http.NewRequest("GET", actionURL, nil)
	req.Header.Add("Authorization", "Bearer "+accesstoken)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, errors.New("Cannot make request")
	}

	if content, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Println(err)
		return nil, errors.New("Cannot read response")
	}

	var res1 = struct {
		TeamDrives []struct {
			Id string `json:"id"`
		} `json:"teamDrives"`
	}{}

	if resp.StatusCode != 200 {
		log.Println(string(content))
		return nil, errors.New("Return status code " + strconv.Itoa(resp.StatusCode))
	}

	if err = json.Unmarshal(content, &res1); err != nil {
		log.Println(err)
		return nil, err
	}

	var res []string
	for _, tid := range res1.TeamDrives {
		res = append(res, tid.Id)
	}

	return res, nil
}

func UploadFile(file string, accesstoken string, metadata interface{}) (*UploadResult, error) {
	var actionURL = GAPIUrl + "/upload/drive/v3/files?uploadType=multipart&supportsTeamDrives=true"

	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	fi, err := fd.Stat()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	var content []byte

	//Build body
	body := new(bytes.Buffer)
	bound := "abcdef-ghik"
	body.WriteString("--" + bound + "\n")
	body.WriteString("Content-Type: application/json; charset=UTF-8\n\n")
	jsonMeta, _ := json.Marshal(metadata)
	body.WriteString(string(jsonMeta) + "\n\n")
	body.WriteString("--" + bound + "\n")
	body.WriteString("Content-Type: image/gif\n\n")
	io.Copy(body, fd)
	body.WriteString("\n--" + bound + "--")

	//log.Println(body)

	//Build request
	req, err := http.NewRequest("POST", actionURL, body)
	req.Header.Add("Authorization", "Bearer "+accesstoken)
	req.Header.Add("Content-Type", "multipart/related; boundary=\""+bound+"\"")
	req.Header.Add("Content-Length", strconv.Itoa(int(fi.Size())))

	//Exec request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if content, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	//log.Println(string(content))

	if resp.StatusCode != 200 {
		return nil, errors.New("Status code with " + strconv.Itoa(resp.StatusCode))
	}

	var res = &UploadResult{}

	if err = json.Unmarshal(content, res); err != nil {
		return nil, err
	}
	return res, nil
}
