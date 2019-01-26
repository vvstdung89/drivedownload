package goutils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)
type Cache struct {
	cache *lru.Cache
	lvdb *leveldb.DB
	persistent bool
}


func Init(name string, size int, persistent bool) *Cache {
	lruCache, _ := lru.New(size)
	lvdb, err := leveldb.OpenFile("levelDB-"+name , nil)
	if err != nil {
		panic(err)
	}
	return &Cache{lruCache,lvdb,persistent}
}

func (self *Cache) SaveCacheData(key string, data interface{},expire int64) error {
	type CacheData struct {
		Data   []byte
		Expire int64
	}

	var databuffer bytes.Buffer
	var cachebuffer bytes.Buffer
	enc := gob.NewEncoder(&databuffer)

	err := enc.Encode(data)
	if err != nil {
		fmt.Println("There was an error:", err)
		return err
	}

	enc = gob.NewEncoder(&cachebuffer)
	err = enc.Encode(CacheData{Data: databuffer.Bytes(), Expire: expire})
	if err != nil {
		return err
	}
	self.cache.Add(key, cachebuffer.Bytes())

	if self.persistent == true {
		self.lvdb.Put([]byte(key), cachebuffer.Bytes(), nil)
	}
	return nil
}

func (self *Cache) GetCacheData(key string, data interface{}) bool {
	var cacheData struct {
		Data   []byte
		Expire int64
	}
	buffer, _ := self.cache.Get(key)
	inMemCache := true
	if buffer == nil {
		if self.persistent == true {
			res,_ := self.lvdb.Get([]byte(key), nil)
			if len(res) != 0 {
				buffer = res
				inMemCache = false
			} else {
				return false
			}
		} else {
			return false
		}
	}

	dec := gob.NewDecoder(bytes.NewReader(buffer.([]byte)))
	err := dec.Decode(&cacheData)
	if err != nil {
		fmt.Println("There was an error:", err)
		return false
	} else {
		if cacheData.Expire > 0 && cacheData.Expire < int64(time.Now().Unix()) {
			return false
		}
		dec := gob.NewDecoder(bytes.NewReader(cacheData.Data))
		err := dec.Decode(data)
		if err != nil {
			fmt.Println("There was an error:", err)
			return false
		}
		if inMemCache == false {
			self.cache.Add(key, cacheData)
		}
		return true
	}
}