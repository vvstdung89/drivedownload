package badger

import (
	"fmt"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"sync"
	"testing"
	"time"
)

func TestStartcUp(t *testing.T) {
	initDB("/tmp/badger")
	err := StartSyncUp("mongodb://127.0.0.1:27017", "ChangeStreamDB", "File", func(d bsonx.Doc) []KVSync {
		return []KVSync{{"id", d.Lookup("drive").StringValue() + "_" + d.Lookup("id").StringValue()}}
	})
	cnt := 0

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Syncing...")
	}

	for {
		time1 := time.Now()
		cnt++
		fmt.Println(time.Since(time1).Seconds())
		//_ = badgerDB.View(func(tx *badger.Txn) error {
		//	time1 := time.Now()
		//	i, _ := tx.Get([]byte("a"))
		//	_ = i.Value(func(b []byte) error {
		//		return nil
		//	})
		//	fmt.Println(time.Since(time1).Seconds())
		//	return nil
		//})

	}

	wa := sync.WaitGroup{}
	wa.Add(1)
	wa.Wait()
}
