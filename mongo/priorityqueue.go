package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type PriorityQueue struct {
	Client *mongo.Client
	Collection *mongo.Collection
	QueueName string
	WorkerName string
}

func NewPriorityQueue(ep string, queueName string, workerName string) *PriorityQueue{
	client, err := mongo.NewClient(options.Client().ApplyURI(ep))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("PriorityQueue").Collection(queueName)

	//set indexing
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys := bson.D{{"pri", -1},{"time_created",1}}
	index := mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	keys = bson.D{{"id", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	index.Options = new(options.IndexOptions).SetUnique(true)

	collection.Indexes().CreateOne(context.Background(), index, opts)
	queue := &PriorityQueue{
		client,
		collection,
		queueName,
		workerName,
	}
	return queue
}

func (s *PriorityQueue) addTask(id, msg string, pri int){
	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{}
	options.SetUpsert(true)

	filter := bson.D{
		{"id",id},
	}
	update := bson.D{
		{"$set", bson.D{
			{"id",id},
			{"msg",msg},
		}},
		{"$setOnInsert", bson.D{
			{"status",0},
			{"time_created",time.Now()},
		}},
		{"$inc", bson.D{{"pri",pri}}},
	}
	s.Collection.FindOneAndUpdate(ctx, filter, update,options)
}

func (s *PriorityQueue) getTask() bson.Raw{
	filter := bson.D{
		{"status",0},
	}
	update := bson.D{
		{"$set", bson.D {
			{"worker", bson.D{
					{s.WorkerName, bson.D{{"time_start", time.Now()}}},
				},
			},
		}},
		{"$inc", bson.D{{"status",1}}},
	}

	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{
		Sort: bson.D{{"pri", -1},{"time_created",1}},
	}
	options.SetReturnDocument(1)

	res:=s.Collection.FindOneAndUpdate(ctx,filter, update, options)
	if res.Err() != nil {
		log.Println(res.Err())
		return nil
	} else {
		doc,_ := res.DecodeBytes()
		return doc
	}
}

func (s *PriorityQueue) setTaskPriority(id string, pri int) error{
	filter := bson.D{
		{"id",id},
	}
	update := bson.D{
		{"$set", bson.D {
			{"pri", pri},
		}},
	}

	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{}
	res:=s.Collection.FindOneAndUpdate(ctx,filter, update, options)
	if res.Err() != nil {
		log.Println(res.Err())
	}
	return res.Err()
}

func (s *PriorityQueue) endTask(id string, success bool, msg string) error{
	filter := bson.D{
		{"id",id},
	}
	status:=-1
	if success == true {
		status=100
	}
	update := bson.D{
		{"$set", bson.D {
			{"status", status},
			{"worker." + s.WorkerName + ".time_end",time.Now()},
			{"worker." + s.WorkerName + ".success",success},
			{"worker." + s.WorkerName + ".msg",msg},},
		},
	}

	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{}
	res:=s.Collection.FindOneAndUpdate(ctx,filter, update, options)
	if res.Err() != nil {
		log.Println(res.Err())
	}
	return res.Err()
}

