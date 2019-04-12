package badger

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/vvstdung89/goutils/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"sync"
	"time"
)

type BadgerDBWrapper struct {
	*badger.DB
}

var badgerDBList = make(map[string]*BadgerDBWrapper)

type KVSync struct {
	key   string
	value string
}

func GetBadgerDB(dbPath string) *BadgerDBWrapper {
	if badgerDBList[dbPath] != nil {
		return badgerDBList[dbPath]
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	var badgerDB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	badgerDBList[dbPath] = &BadgerDBWrapper{badgerDB}
	return badgerDBList[dbPath]
}

func (badgerDB *BadgerDBWrapper) Get(key string) (value interface{}) {
	value = nil
	err := badgerDB.View(func(tx *badger.Txn) error {
		i, err := tx.Get([]byte(key))
		if err != nil {
			return err
		}
		err = i.Value(func(b []byte) error {
			value = b
			return nil
		})
		if err != nil {
			value = nil
		}
		return err
	})
	if err != nil {
		return nil
	}
	return value
}

func (badgerDB *BadgerDBWrapper) StartSyncUp(mongoPath string, database string, col string, getKeyValueFromDoc func(d bsonx.Doc) []KVSync) {

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
}
