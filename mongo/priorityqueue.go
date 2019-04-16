package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"time"
)

type PriorityQueue struct {
	Client     *mongo.Client
	Collection *mongo.Collection
	QueueName  string
	WorkerName string
}

type ItemInfo struct {
	Id     string `json:"id"`
	Msg    string `json:"msg"`
	GetCnt int    `json:"getCnt"`
	Status string `json:"status"`
}

func NewPriorityQueue(ep string, queueName string, workerName string) (*PriorityQueue, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(ep))
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}
	collection := client.Database("PriorityQueue").Collection(queueName)

	//set indexing
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys := bson.D{{"pri", -1}, {"time_created", 1}}
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
	return queue, nil
}

func (s *PriorityQueue) AddTask(id, msg string, pri int) {
	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{}
	options.SetUpsert(true)

	filter := bson.D{
		{"id", id},
	}
	update := bson.D{
		{"$setOnInsert", bson.D{
			{"id", id},
			{"msg", msg},
			{"status", "Init"},
			{"getCnt", 0},
			{"time_created", time.Now()},
		}},
		{"$inc", bson.D{{"pri", pri}}},
	}
	s.Collection.FindOneAndUpdate(ctx, filter, update, options)
}

func (s *PriorityQueue) GetTask() (*ItemInfo, error) {
	filter := bson.D{
		{"getCnt", 0},
	}
	update := bson.D{
		{"$set", bson.M{
			"result.time_start": time.Now(),
			"status":            "Processing",
		}},
		{"$inc", bson.D{{"getCnt", 1}}},
	}

	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{
		Sort: bson.D{{"pri", -1}, {"time_created", 1}},
	}
	options.SetReturnDocument(1)

	doc := s.Collection.FindOneAndUpdate(ctx, filter, update, options)
	if doc.Err() != nil {
		log.Println(doc.Err())
		return nil, doc.Err()
	} else {
		res := &ItemInfo{}
		err := doc.Decode(res)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func (s *PriorityQueue) SetTaskPriority(id string, pri int) error {
	filter := bson.D{
		{"id", id},
	}
	update := bson.D{
		{"$set", bson.D{
			{"pri", pri},
		}},
	}

	ctx := context.Background()
	options := &options.FindOneAndUpdateOptions{}
	res := s.Collection.FindOneAndUpdate(ctx, filter, update, options)
	if res.Err() != nil {
		log.Println(res.Err())
	}
	return res.Err()
}

func (s *PriorityQueue) UpdateField(filter map[string]interface{}, _update map[string]interface{}) error {
	ctx := context.Background()
	options := &options.UpdateOptions{}

	update := bson.M{
		"$set": _update,
	}

	_, err := s.Collection.UpdateMany(ctx, filter, update, options)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *PriorityQueue) AddToSet(filter map[string]interface{}, _update map[string]interface{}) error {
	ctx := context.Background()
	options := &options.UpdateOptions{}

	update := bson.M{
		"$addToSet": _update,
	}

	_, err := s.Collection.UpdateMany(ctx, filter, update, options)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *PriorityQueue) UpdateStatus(filter map[string]interface{}, status string, msg string) error {

	overStatus := status
	if strings.Index(status, "Fail") > 1 {
		overStatus = "Fail"
	}
	update := bson.D{
		{"$set", bson.D{
			{"status", overStatus},
			{"result.time_update", time.Now()},
			{"result.status", status},
			{"result.worker", s.WorkerName},
			{"result.msg", msg}},
		},
	}

	ctx := context.Background()
	options := &options.UpdateOptions{}
	_, err := s.Collection.UpdateMany(ctx, filter, update, options)
	if err != nil {
		log.Println(err)
	}
	return err
}
