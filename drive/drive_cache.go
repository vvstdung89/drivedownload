package drive

import (
	"github.com/vvstdung89/goutils/lrucache"
	"github.com/vvstdung89/goutils/resource_lock"
)

var lockStream *resource_lock.Lock
var lockDown *resource_lock.Lock
var driveStreamCache *lrucache.Cache
var driveDownCache *lrucache.Cache

func init() {
	driveStreamCache = lrucache.Init("drivestream", 1000*1000, true)
	driveDownCache = lrucache.Init("drivedown", 1000*1000, true)
	lockStream = resource_lock.NewResourceLock(100 * 1000)
	lockDown = resource_lock.NewResourceLock(100 * 1000)
}

//get drive stream link with cache
func GetDriveStream(driveID string, accessToken string) DriveStreamInfo {
	lockFile := lockStream.GetResourceLock("stream-" + driveID)
	lockFile.Lock()
	defer lockFile.Unlock()

	var driveStreamInfo DriveStreamInfo
	if isOK := driveStreamCache.GetCacheData("stream-"+driveID, &driveStreamInfo); isOK == true {
		return driveStreamInfo
	}
	driveStreamInfo = StreamInfo(driveID, accessToken)
	driveStreamCache.SaveCacheData("stream-"+driveID, driveStreamInfo, driveStreamInfo.ExpireTime)
	return driveStreamInfo
}

//get drive download link with cache
func GetDriveDownloadLink(driveID string, accessToken string) DriveDownInfo {
	lockFile := lockDown.GetResourceLock("down-" + driveID)
	lockFile.Lock()
	defer lockFile.Unlock()

	var driveDownInfo DriveDownInfo
	if isOK := driveDownCache.GetCacheData("down-"+driveID, &driveDownInfo); isOK == true {
		return driveDownInfo
	}
	driveDownInfo = DownloadInfo(driveID, accessToken)
	driveDownCache.SaveCacheData("down-"+driveID, driveDownInfo, driveDownInfo.ExpireTime)
	return driveDownInfo
}
