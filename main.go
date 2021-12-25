package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	EstimatedTime int64
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

func (client MyClient) GetTask(id string) (TaskDb, bool) {
	var rep TaskDb
	usersCollection := client.Database("testing").Collection("users")
	final, _ := primitive.ObjectIDFromHex(id)
	cur, err := usersCollection.Find(context.TODO(), bson.M{"_id": final})

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.TODO())
	if cur.TryNext(context.TODO()) {
		cur.Decode(&rep)
	} else {
		return rep, false
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return rep, true
}

func (client MyClient) NewTask(data Task) (*mongo.InsertOneResult, error) {
	usersCollection := client.Database("testing").Collection("users")
	index, err := usersCollection.InsertOne(context.TODO(), data)
	//fmt.Println(index)
	return index, err
}

func (client MyClient) RemoveTask(id string) (*mongo.DeleteResult, error) {
	usersCollection := client.Database("testing").Collection("users")
	final, _ := primitive.ObjectIDFromHex(id)
	index, err := usersCollection.DeleteOne(context.TODO(), bson.M{"_id": final})
	//fmt.Println(index, err)
	return index, err
}

var client MyClient

func main() {
	client = Connect()
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/Tasks", GetAllTasks).Methods("GET")
	r.HandleFunc("/Task/{id}", GetTaskID).Methods("GET")
	r.HandleFunc("/Task/{id}", DeleteTaskID).Methods(http.MethodDelete)
	r.HandleFunc("/Task", AddTask).Methods(http.MethodPost)
	// Bind to a port and pass our router in
	fmt.Println("Server is up in 8080")
	http.ListenAndServe(":8080", r)
}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	var exampleBytes []byte
	var err error
	exampleBytes, err = json.Marshal(client.GetAllTasks())
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(exampleBytes)
	}
}

func GetTaskID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var exampleBytes []byte
	var err error
	rep, count := client.GetTask(vars["id"])
	exampleBytes, err = json.Marshal(rep)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "application/json")
		if count {
			w.Write(exampleBytes)
		} else {
			w.Write([]byte("[]"))
		}
	}
}

func DeleteTaskID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rep, err := client.RemoveTask(vars["id"])
	if rep.DeletedCount > 0 && err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var anyJson map[string]interface{}
	err := decoder.Decode(&anyJson)
	fmt.Println(err)
	var task Task
	if err := TryCatch(func() {
		task.Title = anyJson["Title"].(string)
		task.DateStart, _ = dateparse.ParseLocal(anyJson["DateStart"].(string))
		task.DateStop, _ = dateparse.ParseLocal(anyJson["DateStop"].(string))
		timeestim, _ := dateparse.ParseLocal(anyJson["EstimatedTime"].(string))
		task.EstimatedTime = time.Now().UnixNano() - timeestim.UnixNano()
		task.Status = Status(anyJson["Status"].(string))
		task.Tag = anyJson["Tag"].(string)
	})(); err != nil {
		fmt.Println("erreur")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	index, errbdd := client.NewTask(task)
	if errbdd == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{\"ID\":\"" + index.InsertedID.(primitive.ObjectID).Hex() + "\"}"))
	} else {
		w.WriteHeader(http.StatusNotModified)
	}

}

func DispHelp() []byte {
	return []byte(`    {
        "Title": "mon titre",
        "DateStart": "jj/mm/aaaa hh:mm:ss",
        "DateStop": "jj/mm/aaaa hh:mm:ss",
        "EstimatedTime": "jj/mm/aaaa hh:mm:ss",
        "Status": "Progress or Done or ToDo or Failed",
        "Tag": "mon tag"
    }`)
}

func TryCatch(f func()) func() error {
	return func() (err error) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
				return
			}
		}()
		f() // calling the decorated function
		return err
	}
}

//client.NewTask(Task{Title: "tttt", DateStart: time.Date(2020, time.April,
//	11, 21, 34, 01, 0, time.UTC), DateStop: time.Date(2020, time.April,
//	11, 21, 34, 01, 0, time.UTC), EstimatedTime: time.Duration.Hours(1), Status: Status(Progress), Tag: "mon tagg"})
// client.RemoveTask("tttt")
