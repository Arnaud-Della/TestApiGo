package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Status string

const (
	Done     = "Done"
	Progress = "Progress"
	ToDo     = "ToDo"
	Failed   = "Failed"
)

type MyClient struct {
	*mongo.Client
}

type TaskDb struct {
	ID            string    `bson:"_id"`
	Title         string    `bson:"title"`
	DateStart     time.Time `bson:"dateStart"`
	DateStop      time.Time `bson:"dateStop"`
	EstimatedTime float64   `bson:"estimatedTime"`
	Status        Status    `bson:"status"`
	Tag           string    `bson:"tag"`
}

type Task struct {
	Title         string
	DateStart     time.Time
	DateStop      time.Time
	EstimatedTime float64
	Status        Status
	Tag           string
}

func Connect() MyClient {
	clientOptions := options.Client().ApplyURI("mongodb://Arnaud:pass@localhost/testing")
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return MyClient{client}
}

func (client MyClient) GetAllTasks() []TaskDb {
	tab := []TaskDb{}
	usersCollection := client.Database("testing").Collection("users")
	cur, err := usersCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		//print element data from collection
		var a TaskDb
		cur.Decode(&a)
		tab = append(tab, a)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return tab
}

func (client MyClient) NewTask(data Task) (*mongo.InsertOneResult, error) {
	usersCollection := client.Database("testing").Collection("users")
	index, err := usersCollection.InsertOne(context.TODO(), data)
	//fmt.Println(index)
	return index, err
}

func (client MyClient) RemoveTask(data string) (*mongo.DeleteResult, error) {
	usersCollection := client.Database("testing").Collection("users")
	index, err := usersCollection.DeleteOne(context.TODO(), bson.M{"title": data})
	//fmt.Println(index, err)
	return index, err
}

func main() {
	fmt.Println("Lançement du programme ...")
	client := Connect()
	a := client.GetAllTasks()
	for _, item := range a {
		fmt.Println(item)
	}
	//client.NewTask(Task{Title: "tttt", DateStart: time.Date(2020, time.April,
	//	11, 21, 34, 01, 0, time.UTC), DateStop: time.Date(2020, time.April,
	//	11, 21, 34, 01, 0, time.UTC), EstimatedTime: time.Duration.Hours(1), Status: Status(Progress), Tag: "mon tagg"})
	client.RemoveTask("tttt")
}
