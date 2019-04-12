package badger

import (
	"fmt"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"sync"
	"testing"
)

func TestStartcUp(t *testing.T) {
	db := GetBadgerDB("/tmp/badger")
	db.StartSyncUp("mongodb://127.0.0.1:27017", "ChangeStreamDB", "File", func(d bsonx.Doc) []KVSync {
		return []KVSync{{d.Lookup("id").StringValue(), d.Lookup("drive").StringValue() + "_" + d.Lookup("id").StringValue()}}
	})
	x := db.Get("R6RZxOMTtkIMR9T3yEXk")
	if x != nil {
		fmt.Println(string(x.([]byte)))
	}
	wa := sync.WaitGroup{}
	wa.Add(1)
	wa.Wait()
}
