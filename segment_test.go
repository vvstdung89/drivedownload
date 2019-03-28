package goutils

import (
	"log"
	"testing"
)

func TestSegmentVideo(*testing.T) {
	var src = "/tmp/filetmp"
	var dst = "/tmp/filetmp_hls.m3u8"
	var targetDuration = 6
	var maxSize = 10 * 1024 * 1024
	var prefixFile = "./pixel.gif"
	seg := SegmentVideo(src, targetDuration, maxSize, dst, prefixFile)
	log.Println(seg)
}
