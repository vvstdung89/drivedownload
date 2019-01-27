package lrucache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

type CacheData struct {
	Data   interface{}
	Expire int64
}

type Cache struct {
	cache      *lru.Cache
	lvdb       *leveldb.DB
	persistent bool
}

func Init(name string, size int, persistent bool) *Cache {
	lruCache, _ := lru.New(size)
	lvdb, err := leveldb.OpenFile("levelDB-"+name, nil)
	if err != nil {
		panic(err)
	}
	//remove old item
	//go func() {
	//	for _ = range time.NewTicker(time.Minute * 30).C {
	//		iter := lvdb.NewIterator(util.BytesPrefix([]byte("-")), nil)
	//		for iter.Next() {
	//			var cacheData CacheData
	//			dec := gob.NewDecoder(bytes.NewReader(iter.Value()))
	//			err = dec.Decode(&cacheData)
	//			if cacheData.Expire <= time.Now().Add(-6*time.Hour).Unix() {
	//				lvdb.Delete(iter.Key(), nil)
	//			}
	//		}
	//		iter.Release()
	//	}
	//}()
	return &Cache{lruCache, lvdb, persistent}
}

func (self *Cache) Remove(key string) {
	self.cache.Remove(key)
	if self.persistent == true {
		self.lvdb.Delete([]byte(key), nil)
	}
}

func (self *Cache) SaveCacheData(key string, data interface{}, expire int64) error {
	self.cache.Add(key, CacheData{Data: data, Expire: expire})
	if self.persistent == true && expire == 0 {
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
		self.lvdb.Put([]byte(key), cachebuffer.Bytes(), nil)
	}
	return nil
}

func (self *Cache) GetCacheData(key string, data interface{}) bool {
	var cacheData CacheData
	buffer, _ := self.cache.Get(key)

	if buffer == nil {
		if self.persistent == true {
			res, _ := self.lvdb.Get([]byte(key), nil)
			if len(res) != 0 {
				buffer = res
			} else {
				return false
			}
		} else {
			return false
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
			dec := gob.NewDecoder(bytes.NewReader(cacheData.Data.([]byte)))
			err := dec.Decode(data)
			if err != nil {
				fmt.Println("There was an error:", err)
				return false
			}

			self.cache.Add(key, cacheData)

			return true
		}

	} else {
		data = buffer.(CacheData).Data
		return true
	}

}
