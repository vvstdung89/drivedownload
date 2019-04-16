package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

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

func (s *MongoClient) GetCollection(db, col string) *mongo.Collection {
	collection := s.Client.Database(db).Collection(col)
	return collection
}

func (s *MongoClient) CreateIndex(db, col string, keys bson.M, unique bool) error {
	collection := s.Client.Database(db).Collection(col)
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	index := mongo.IndexModel{}
	index.Keys = keys
	index.Options = &options.IndexOptions{Unique: &unique}

	_, err := collection.Indexes().CreateOne(context.Background(), index, opts)

	return err
}
