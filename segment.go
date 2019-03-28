package goutils

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SegmentInfo struct {
	p int        //prefix size
	d int        //target duration
	s [][]int    //segment size
	t [][]string //segment time
}

func SegmentVideo(src string, targetDuration int, maxSize int, dst string, prefixFile string) *SegmentInfo {

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

	prefixReader, err := os.Open(prefixFile)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer prefixReader.Close()

	var segFile = 0
	var fileSize = 0
	var size = 0
	var curSegSize = 0

	prefixInfo, _ := prefixReader.Stat()

	var segmentInfo = SegmentInfo{
		d: 0,
		t: [][]string{},
		s: [][]int{},
		p: int(prefixInfo.Size()),
	}
	segmentInfo.s = append(segmentInfo.s, []int{})
	segmentInfo.t = append(segmentInfo.t, []string{})

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

			if segmentInfo.s[segFile] == nil {
				segmentInfo.s[segFile] = []int{}
			}
			segmentInfo.s[segFile] = append(segmentInfo.s[segFile], size)

			if curSegSize > maxSize {

				to, err := os.OpenFile(src+"_"+strconv.Itoa(segFile), os.O_RDWR|os.O_CREATE, 0666)
				if err != nil {
					log.Fatal(err)
				}
				to.Truncate(0)
				defer to.Close()

				_, err = io.Copy(to, prefixReader)
				if err != nil {
					log.Fatal(err)
					return nil
				}
				prefixReader.Seek(0, 0)
				if err := exec.Command("dd", "bs=1M", "if="+src+".ts", "skip="+strconv.Itoa(fileSize+size-curSegSize), "count="+strconv.Itoa(curSegSize), "iflag=skip_bytes,count_bytes", "of="+src+"_"+strconv.Itoa(segFile), "oflag=append", "conv=notrunc").Run(); err != nil {
					log.Printf("Command finished with error: %v", err)
				}

				segFile++
				curSegSize = 0

				if len(segmentInfo.s) < segFile+1 {
					segmentInfo.s = append(segmentInfo.s, []int{})
				}
				if len(segmentInfo.t) < segFile+1 {
					segmentInfo.t = append(segmentInfo.t, []string{})
				}
			}
		}

		if strings.HasPrefix(line, "#EXT-X-TARGETDURATION") {
			regex := regexp.MustCompile(":([\\d]+)")
			match := regex.FindStringSubmatch(line)
			segmentInfo.d, _ = strconv.Atoi(match[1])
		}

		if strings.HasPrefix(line, "#EXTINF") {
			regex := regexp.MustCompile(":(.+),")
			match := regex.FindStringSubmatch(line)
			segmentInfo.t[segFile] = append(segmentInfo.t[segFile], match[1])
		}

		if strings.HasPrefix(line, "#EXT-X-ENDLIST") {
			if curSegSize > 0 {
				to, err := os.OpenFile(src+"_"+strconv.Itoa(segFile), os.O_RDWR|os.O_CREATE, 0666)
				if err != nil {
					log.Fatal(err)
				}
				to.Truncate(0)
				defer to.Close()

				_, err = io.Copy(to, prefixReader)
				if err != nil {
					log.Fatal(err)
					return nil
				}

				if err := exec.Command("dd", "bs=1M", "if="+src+".ts", "skip="+strconv.Itoa(fileSize+size-curSegSize), "count="+strconv.Itoa(curSegSize), "iflag=skip_bytes,count_bytes", "of="+src+"_"+strconv.Itoa(segFile), "oflag=append", "conv=notrunc").Run(); err != nil {
					log.Printf("Command finished with error: %v", err)
				}
			}
			break
		}
	}

	return &segmentInfo
}
