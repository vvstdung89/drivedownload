package goutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)


type Jar struct {
	lk      sync.Mutex
	cookies []*http.Cookie
}

func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = []*http.Cookie{}
	return jar
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies = cookies
	jar.lk.Unlock()
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

func ChunkDownload(dst string, streamLink string, cookie string, size int, chunk int, parallel int) error {
	readWriteTimeout := time.Minute * 10
	connectTimeout := time.Second * 30

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
	accummulate := make(chan int, taskNum)
	var wg sync.WaitGroup
	for i := 0; i < taskNum; i++ {
		job <- i
		wg.Add(1)
	}

	cont := true
	for i := 0; i < parallel; i++ {
		go func() {
			for true {
				if len(job) == 0 {
					break
				}
				taskID := <-job
				if cont == false {
					wg.Done()
					continue
				}

				try := 0
				startRange := taskID * chunk
				endRange := ((taskID + 1) * chunk) - 1
				if taskID == taskNum-1 {
					endRange = size
				}

				for cont == true {
					u, err := url.Parse(streamLink)
					cookies := []*http.Cookie{}
					cookies = append(cookies, &http.Cookie{Name:"DRIVE_STREAM",Value: cookie})
					client := &http.Client{
						Jar: NewJar(),
						Timeout: readWriteTimeout,
						Transport: &http.Transport{
							ResponseHeaderTimeout: connectTimeout,
						},
					}
					client.Jar.SetCookies(u, cookies)

					start := time.Now()
					req, _ := http.NewRequest("GET", streamLink, nil)
					range_header := "bytes=" + strconv.Itoa(startRange) + "-" + strconv.Itoa(endRange)
					req.Header.Add("Range", range_header)
					resp, err := client.Do(req)

					defer func(){
						if resp != nil {
							resp.Body.Close()
						}
					}()
					elapsed := time.Since(start)

					log.Println("job", taskID, startRange, endRange, cont, elapsed)

					if err == nil {
						if resp == nil {
							cont = false
							wg.Done()
							break
						}
						reader, err1 := ioutil.ReadAll(resp.Body)
						if err1==nil && reader != nil && len(reader) > 0 && (resp.StatusCode == 206 || resp.StatusCode == 200 || resp.StatusCode == 302) {
							go func(){
								if f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0777);err !=nil {
									log.Println(err)
									cont = false
									wg.Done()
									return
								} else if _, err := f.WriteAt([]byte(string(reader)), int64(startRange)); err != nil {
									log.Println(err)
									cont = false
									wg.Done()
									return
								}
								accummulate <- taskID
								wg.Done()
							}()
							break;
						} else {
							err = errors.New("Response error " + strconv.Itoa(resp.StatusCode))
						}
					}

					if err != nil {
						fmt.Println(taskID, err)
						try++
						if try >= 3 {
							cont = false
							wg.Done()
							break
						}
					}
				}


			}
		}()
	}

	wg.Wait()
	fmt.Println(len(accummulate))
	if len(accummulate) != taskNum {
		os.RemoveAll(dst)
		return errors.New("Download fail")
	}



	return nil
}
