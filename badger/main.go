package badger

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/vvstdung89/goutils/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"sync"
	"time"
)

var badgerDB *badger.DB

type KVSync struct {
	key   string
	value string
}

func initDB(dbPath string) {
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	var err error
	badgerDB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
}

func StartSyncUp(mongoPath string, database string, col string, getKeyValueFromDoc func(d bsonx.Doc) []KVSync) error {
	if badgerDB == nil {
		return errors.New("BadgerDB not yet initiated")
	}
	mongoClient := mongo.NewMongoClient(mongoPath)
	docs := mongoClient.RetrieveOplog(database, col)
	var kvsLock sync.Mutex
	var kvs []KVSync

	go func() {
		for {
			d := <-docs
			if d.OperationType == "insert" {
				kvsLock.Lock()
				kvs = append(kvs, getKeyValueFromDoc(d.FullDocument)...)
				kvsLock.Unlock()
			}
			//TODO: delete|update ???
		}
	}()

	go func() {
		ticker := time.Tick(time.Millisecond * 500)
		for {
			<-ticker
			kvsLock.Lock()
			if len(kvs) > 0 {
				if err := badgerDB.Update(func(txn *badger.Txn) error {
					for _, v := range kvs {
						fmt.Println("set", v.key, "to", v.value)
						txn.Set([]byte(v.key), []byte(v.value))
					}
					return nil
				}); err != nil {
					log.Println("Update error", err)
				}
				kvs = []KVSync{}
			}
			kvsLock.Unlock()
		}
	}()
	return nil
}
