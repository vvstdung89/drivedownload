package goutils

import (
	"log"
	"testing"
)

var access_token = ""

func TestRefreshToken(t *testing.T) {
	var refresh_token = "1/_DSrOXBqnr-8IcTKpU0eCRUUG6GNY2FkmEPoFQjwdpfjbZTI-vfMB3ZnJITUr6xO"
	var client_id = "893623223629-42v9e7cvr71l9jt66od6k74rljv0ab1m.apps.googleusercontent.com"
	var secret = ""

	newToken, err := RefreshToken(refresh_token, client_id, secret)
	log.Println(newToken, err)
}

func TestUploadFile(t *testing.T) {
	var file = "./pixel.gif"
	var metadata = UploadMetaData{
		"abcd",
		[]string{"0AM4A9MV0sXjAUk9PVA"},
		"0AM4A9MV0sXjAUk9PVA",
	}

	res, err := UploadFile(file, access_token, metadata)
	log.Println(res, err)
}

func TestGetTeamDrive(t *testing.T) {
	res, err := GetTeamDrive(access_token)
	log.Println(res, err)
}
