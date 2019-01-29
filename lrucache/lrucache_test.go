package lrucache_test

import (
	"github.com/goutils/lrucache"
	"log"
	"testing"
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

	b := abc{}
	isOK := cache.GetCacheData("1", &b)
	if isOK == true {
		log.Println(b)
	} else {
		log.Println("Not in cache")
	}

}
