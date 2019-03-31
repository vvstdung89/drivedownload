package goutils

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SegmentInfo struct {
	D int        //target duration
	S [][]int    //segment size
	T [][]string //segment time
}

func SegmentVideo(src string, targetDuration int, maxSize int) *SegmentInfo {

	//transcode to ffmpeg
	if err := exec.Command("ffmpeg", "-threads", "1", "-i", src, "-c:a", "copy", "-c:v", "copy", "-hls_playlist_type", "vod", "-hls_time", strconv.Itoa(targetDuration), "-hls_flags", "single_file", "-threads", "1", src+".m3u8").Run(); err != nil {
		log.Fatal(err)
		return nil
	}

	//copy segment
	rfile, err := os.Open(src + ".m3u8")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer rfile.Close()

	var segFile = 0
	var fileSize = 0
	var size = 0
	var curSegSize = 0

	var segmentInfo = SegmentInfo{
		D: 0,
		T: [][]string{},
		S: [][]int{},
	}
	segmentInfo.S = append(segmentInfo.S, []int{})
	segmentInfo.T = append(segmentInfo.T, []string{})

	scanner := bufio.NewScanner(rfile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#EXT-X-BYTERANGE") {
			regex := regexp.MustCompile("([\\d]+)@([\\d]+)")
			match := regex.FindStringSubmatch(line)

			start := curSegSize
			size, _ = strconv.Atoi(match[1])
			fileSize, _ = strconv.Atoi(match[2])
			curSegSize = start + size

			if segmentInfo.S[segFile] == nil {
				segmentInfo.S[segFile] = []int{}
			}
			segmentInfo.S[segFile] = append(segmentInfo.S[segFile], size)

			if curSegSize > maxSize {

				to, err := os.OpenFile(src+"_"+strconv.Itoa(segFile), os.O_RDWR|os.O_CREATE, 0666)
				if err != nil {
					log.Fatal(err)
				}
				to.Truncate(0)
				defer to.Close()

				if err := exec.Command("dd", "bs=1M", "if="+src+".ts", "skip="+strconv.Itoa(fileSize+size-curSegSize), "count="+strconv.Itoa(curSegSize), "iflag=skip_bytes,count_bytes", "of="+src+"_"+strconv.Itoa(segFile)).Run(); err != nil {
					log.Printf("Command finished with error: %v", err)
				}

				segFile++
				curSegSize = 0

				if len(segmentInfo.S) < segFile+1 {
					segmentInfo.S = append(segmentInfo.S, []int{})
				}
				if len(segmentInfo.T) < segFile+1 {
					segmentInfo.T = append(segmentInfo.T, []string{})
				}
			}
		}

		if strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {
			regex := regexp.MustCompile(":([\\d]+)")
			match := regex.FindStringSubmatch(line)
			segmentInfo.D, _ = strconv.Atoi(match[1])
		}

		if strings.HasPrefix(line, "#EXTINF") {
			regex := regexp.MustCompile(":(.+),")
			match := regex.FindStringSubmatch(line)
			segmentInfo.T[segFile] = append(segmentInfo.T[segFile], match[1])
		}

		if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			if curSegSize > 0 {
				to, err := os.OpenFile(src+"_"+strconv.Itoa(segFile), os.O_RDWR|os.O_CREATE, 0666)
				if err != nil {
					log.Fatal(err)
				}
				to.Truncate(0)
				defer to.Close()

				if err := exec.Command("dd", "bs=1M", "if="+src+".ts", "skip="+strconv.Itoa(fileSize+size-curSegSize), "count="+strconv.Itoa(curSegSize), "iflag=skip_bytes,count_bytes", "of="+src+"_"+strconv.Itoa(segFile)).Run(); err != nil {
					log.Printf("Command finished with error: %v", err)
				}
			}
			break
		}
	}

	return &segmentInfo
}
