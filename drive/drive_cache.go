package drive

import (
	"github.com/vvstdung89/goutils/Lrucache"
	"github.com/vvstdung89/goutils/resource_lock"
	"sync"
)

var lockStream = &sync.Mutex{}
var lockDown = &sync.Mutex{}
var driveStreamCache *lrucache.Cache
var driveDownCache *lrucache.Cache

func init() {
	driveStreamCache = lrucache.Init("drivestream", 1000*1000, false)
	driveDownCache = lrucache.Init("drivedown", 1000*1000, false)
}

//get drive stream link with cache
func GetDriveStream(driveID string, accessToken string) DriveStreamInfo {
	lockFile := resource_lock.NewResourceLock(100 * 1000).GetResourceLock("stream-" + driveID)
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
	lockFile := resource_lock.NewResourceLock(100 * 1000).GetResourceLock("down-" + driveID)
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
