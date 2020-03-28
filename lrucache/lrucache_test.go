package lrucache_test

import (
	"fmt"
	"github.com/vvstdung89/goutils/lrucache"
	"log"
	"testing"
	"time"
)

type abc struct {
	A string
	C []int
}

func TestCache_GetCacheData(t *testing.T) {
	a := abc{}
	a.A = "111"
	a.C = []int{1, 2, 3}

	cache := lrucache.Init("a", 1, true)
	if isOK := cache.SaveCacheData("1", &a, 0); isOK != nil {
		panic(isOK)
	}

	if isOK := cache.SaveCacheData("2", a.A, 0); isOK != nil {
		panic(isOK)
	}

	time1 := time.Now()
	xxx := abc{}
	bb, isOK := cache.GetCacheData("1", &xxx)
	b := bb.(*abc)
	fmt.Println(time.Since(time1).Seconds())
	if isOK == true {
		log.Println(b)
		log.Println(xxx)
	} else {
		log.Println("Not in cache")
	}

	time2 := time.Now()
	cc, isOK := cache.GetCacheData("1", &abc{})
	c := cc.(*abc)
	fmt.Println(time.Since(time2).Seconds())
	if isOK == true {
		log.Println(c)
	} else {
		log.Println("Not in cache")
	}
}
