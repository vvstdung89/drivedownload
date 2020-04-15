package drive

import (
	"encoding/gob"
	"github.com/vvstdung89/goutils/lrucache"
	"github.com/vvstdung89/goutils/resource_lock"
)

var lockStream *resource_lock.Lock
var lockDown *resource_lock.Lock
var driveStreamCache *lrucache.Cache
var driveDownCache *lrucache.Cache
var driveResCache *lrucache.Cache

func init() {
	driveStreamCache = lrucache.Init("drivestream", 1000*1000, true)
	driveResCache = lrucache.Init("driveresolution", 1000*1000, true)
	driveDownCache = lrucache.Init("drivedown", 1000*1000, true)
	lockStream = resource_lock.NewResourceLock(1000 * 1000)
	lockDown = resource_lock.NewResourceLock(1000 * 1000)
	gob.Register(DriveStreamInfo{})
	gob.Register(DriveDownInfo{})
	gob.Register([]string{})
}

//get drive stream link with cache
func GetDriveStream(driveID string, accessToken string) DriveStreamInfo {
	lockFile := lockStream.GetResourceLock("stream-" + driveID)
	lockFile.Lock()
	defer lockFile.Unlock()

	driveStreamInfo, isOK := driveStreamCache.GetCacheData("stream-" + driveID)
	if isOK == true {
		return driveStreamInfo.(DriveStreamInfo)
	}
	driveStreamInfo = StreamInfo(driveID, accessToken)
	driveStreamCache.SaveCacheData("stream-"+driveID, driveStreamInfo, driveStreamInfo.(DriveStreamInfo).ExpireTime)

	if len(driveStreamInfo.(DriveStreamInfo).Streams) > 0 {
		resArr := []string{}
		for res, _ := range driveStreamInfo.(DriveStreamInfo).Streams {
			resArr = append(resArr, res)
		}
		driveResCache.SaveCacheData("res-"+driveID, resArr, 0)
	}

	return driveStreamInfo.(DriveStreamInfo)
}

//get drive stream link with cache
func GetDriveRes(driveID string, accessToken string) (resArr []string) {
	lockFile := lockStream.GetResourceLock("res-" + driveID)
	lockFile.Lock()
	defer lockFile.Unlock()

	driveResInfo, isOK := driveResCache.GetCacheData("res-" + driveID)
	if isOK == true {
		return driveResInfo.([]string)
	}

	driveStreamInfo := GetDriveStream(driveID, accessToken)
	for res, _ := range driveStreamInfo.Streams {
		resArr = append(resArr, res)
	}

	driveResCache.SaveCacheData("res-"+driveID, resArr, 0)
	return resArr
}

//get drive download link with cache
func GetDriveDownloadLink(driveID string, accessToken string) DriveDownInfo {
	lockFile := lockDown.GetResourceLock("down-" + driveID)
	lockFile.Lock()
	defer lockFile.Unlock()

	driveDownInfo, isOK := driveDownCache.GetCacheData("down-" + driveID)
	if isOK == true {
		return driveDownInfo.(DriveDownInfo)
	}
	driveDownInfo = DownloadInfo(driveID, accessToken)
	driveDownCache.SaveCacheData("down-"+driveID, driveDownInfo, driveDownInfo.(DriveDownInfo).ExpireTime)

	return driveDownInfo.(DriveDownInfo)
}

//get drive download link with cache
func RemoveDriveStream(driveID string) {
	driveStreamCache.Remove("stream-" + driveID)
}

//get drive download link with cache
func GetDriveDownloadLinkAsync(driveID string, accessToken string) {
	_, isOK := driveDownCache.GetCacheData("down-" + driveID)
	if isOK == true {
		return
	}
	go GetDriveDownloadLink(driveID, accessToken)
	return
}
