package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	_ "go.mongodb.org/mongo-driver/x/bsonx"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type MongoClient struct {
	Client *mongo.Client
}

type CSElem struct {
	ID            bsonx.Doc `json:"id" bson:"_id"`
	OperationType string    `json:"operationType" bson:"operationType"`
	FullDocument  bsonx.Doc `json:"fullDocument" bson:"fullDocument"`
	NS            bsonx.Doc `json:"ns" bson:"ns"`
}

func NewMongoClient(ep string) *MongoClient {
	client, err := mongo.NewClient(options.Client().ApplyURI(ep))
	if err != nil {
		log.Fatal("cannot connect 1", err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal("cannot connect 2", err)
	}

	return &MongoClient{
		Client: client,
	}
}

func (s *MongoClient) getFirstOplogTime(ns string) *primitive.Timestamp {
	collection := s.Client.Database("local").Collection("oplog.rs")
	ctx := context.Background()
	filter := bson.D{bson.E{"ns", ns}}
	res := collection.FindOne(ctx, filter)

	type OplogRes struct {
		Ts primitive.Timestamp `bson:"ts"`
	}
	e := &OplogRes{}
	res.Decode(e)
	return &e.Ts
}

func (s *MongoClient) RetrieveOplog(db string, col string) chan CSElem {
	collection := s.Client.Database(db).Collection(col)
	ctx := context.Background()
	docs := make(chan CSElem)
	var lastToken bsonx.Doc
	resumeTokenPath := "./resume_token_" + db + "_" + col
	//Read and Write token
	f, err := os.OpenFile(resumeTokenPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	resumeToken, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	f.Close()
	go func() {
		for { //save token from 5 seconds ago
			tmpToken := lastToken
			time.Sleep(5 * time.Second)
			if tmpToken == nil {
				continue
			}
			bs, _ := bson.Marshal(tmpToken)
			ioutil.WriteFile(resumeTokenPath, bs, 0777)
		}
	}()

	pipeline := []bson.M{}
	options := &options.ChangeStreamOptions{}

	if resumeToken != nil && len(resumeToken) > 0 {
		rt := &bsonx.Doc{}
		err := bson.Unmarshal(resumeToken, rt)
		if err != nil {
			log.Println(err)
		}
		options.SetResumeAfter(rt)
		lastToken = *rt
	} else {
		options.SetStartAtOperationTime(s.getFirstOplogTime(db + "." + col))
	}

	go func() {
		for {
			cur, err := collection.Watch(ctx, pipeline, options)
			if err != nil {
				log.Println("connect error", err)
				time.Sleep(time.Second * 5)
				continue
			}
			defer cur.Close(ctx)
			for cur.Next(ctx) {
				elem := CSElem{}
				if err := cur.Decode(&elem); err != nil {
					log.Fatal(err)
				}
				lastToken = elem.ID
				docs <- elem
			}

			if err := cur.Err(); err != nil {
				log.Println("watch error", err)
			}
			time.Sleep(time.Second * 10)
		}

	}()

	return docs
}
