package goutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func ChunkDownload(dst string, url string, cookie string, size int, chunk int, parallel int) error {
	readWriteTimeout := time.Minute * 10
	connectTimeout := time.Second * 1

	//set default
	if chunk == 0 {
		chunk = 2000000
	}

	if parallel == 0 {
		parallel = 8
	}

	//calculate task
	taskNum := int(math.Ceil(float64(size) / float64(chunk)))
	job := make(chan int, taskNum)
	accummulate := make(chan string, taskNum)
	for i := 0; i < taskNum; i++ {
		job <- i
	}

	cont := true
	var wg sync.WaitGroup

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			for cont == true {
				taskID := <-job
				var try = 0
				startRange := taskID * chunk
				endRange := ((taskID + 1) * chunk) - 1
				if taskID == taskNum-1 {
					endRange = size
				}

				fileName := dst + "_" + strconv.Itoa(taskID)
				for cont == true {

					client := &http.Client{
						Timeout: readWriteTimeout,
						Transport: &http.Transport{
							ResponseHeaderTimeout: connectTimeout,
						},
					}
					log.Println("job", taskID, startRange, endRange, cont)

					req, _ := http.NewRequest("GET", url, nil)
					range_header := "bytes=" + strconv.Itoa(startRange) + "-" + strconv.Itoa(endRange) // Add the data for the Range header of the form "bytes=0-100"
					req.Header.Add("Range", range_header)
					req.Header.Add("Cookie", cookie)
					resp, _ := client.Do(req)
					defer func() {
						if resp != nil {
							resp.Body.Close()
						}
					}()

					if resp == nil {
						cont = false
						break
					}

					reader, err := ioutil.ReadAll(resp.Body)
					if err == nil {
						err = ioutil.WriteFile(fileName, []byte(string(reader)), 0777) // Write to the file i as a byte array
					}

					if err != nil {
						fmt.Println(err)
						try++
						if try > 3 {
							cont = false
							break
						}
					} else {
						//log.Println(fileName, dst, range_header)
						if err := exec.Command("dd", "if="+fileName, "of="+dst, "bs=100000", "seek="+strconv.Itoa(startRange/100000), "conv=notrunc").Run(); err != nil {
							log.Printf("Command finished with error: %v", err)
							cont = false
						} else {
							accummulate <- dst + "_" + strconv.Itoa(taskID)
						}
						break
					}

				}

				if len(job) == 0 {
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if len(accummulate) != taskNum {
		os.RemoveAll(dst)
		return errors.New("Download fail")
	}

	fmt.Println(len(accummulate))
	wroteFileNum := len(accummulate)
	for i := 0; i < wroteFileNum; i++ {
		file := <-accummulate
		//fmt.Println("Remove file", file)
		os.RemoveAll(file)
	}

	return nil
}
