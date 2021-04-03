package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"server/models"
	"server/util/config"
	"server/util/logger"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database Name
const dbName = "gotodo"

// Collection name
const collName = "todolist"

// collection object/instance
var collection *mongo.Collection

var log = logger.GetLogger()

// create connection with mongo db
func init() {

	//get connection string
	var connStr = config.ViperEnvVariable("MONGODB_CONN")

	// Set client options
	clientOptions := options.Client().ApplyURI(connStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Msg("Connected to MongoDB!")

	collection = client.Database(dbName).Collection(collName)

	log.Info().Msg("Collection instance created!")
}

// GracefulShutdown gracefully exits
func GracefulShutdown(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	timer1 := time.NewTimer(time.Second)
	start := time.Now()

	go func(sinceStart time.Time) {
		<-timer1.C
		elapsed := time.Since(sinceStart)
		//log.Info().Dur("duration", elapsed).Msg("Exiting!")
		log.Info().Msgf("Exiting!, Duration: %s", elapsed)
		os.Exit(0)
	}(start)

	json.NewEncoder(w).Encode("Sutting down gracefully.")
	elapsed := time.Since(start)
	//log.Info().Dur("duration", elapsed).Msg("Composed Shutdown Reply!")
	log.Info().Msgf("Composed Shutdown Reply!, Duration: %s", elapsed)
}

// GetAllTask get all the task route
func GetAllTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	payload := getAllTask()
	json.NewEncoder(w).Encode(payload)
}

// CreateTask create task route
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	var task models.ToDoList
	_ = json.NewDecoder(r.Body).Decode(&task)
	insertOneTask(task)
	json.NewEncoder(w).Encode(task)
}

// TaskComplete update task route
func TaskComplete(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	taskComplete(params["id"])
	json.NewEncoder(w).Encode(params["id"])
}

// UndoTask undo the complete task route
func UndoTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	undoTask(params["id"])
	json.NewEncoder(w).Encode(params["id"])
}

// DeleteTask delete one task route
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	params := mux.Vars(r)
	deleteOneTask(params["id"])
	json.NewEncoder(w).Encode(params["id"])
	// json.NewEncoder(w).Encode("Task not found")

}

// DeleteAllTask delete all tasks route
func DeleteAllTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	count := deleteAllTask()
	json.NewEncoder(w).Encode(count)
}

// get all task from the DB and return it
func getAllTask() []primitive.M {
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	var results []primitive.M
	for cur.Next(context.Background()) {
		var result bson.M
		e := cur.Decode(&result)
		if e != nil {
			log.Fatal().Err(e).Msg("")
		}
		results = append(results, result)

	}

	if err := cur.Err(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	cur.Close(context.Background())
	log.Debug().Msg("Get all tasks")
	return results
}

// Insert one task in the DB
func insertOneTask(task models.ToDoList) {
	insertResult, err := collection.InsertOne(context.Background(), task)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msgf("Inserted a Single Record %s", insertResult.InsertedID)
}

// task complete method, update task's status to true
func taskComplete(task string) {
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": true}}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Debug().Msgf("Task %s complete, count: %d", task, result.ModifiedCount)
}

// task undo method, update task's status to false
func undoTask(task string) {
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": false}}
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msgf("Task %s undo, count: %d", task, result.ModifiedCount)
}

// delete one task from the DB, delete by ID
func deleteOneTask(task string) {
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	d, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msgf("Task %s delete, count: %d", task, d.DeletedCount)
}

// delete all the tasks from the DB
func deleteAllTask() int64 {
	d, err := collection.DeleteMany(context.Background(), bson.D{{}}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msgf("Deleted Document, count: %d", d.DeletedCount)
	return d.DeletedCount
}
