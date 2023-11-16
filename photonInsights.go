package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"photoninsights/datagen"
)

var MongoClient *mongo.Client

var ModulesColl *mongo.Collection
var MessagesColl *mongo.Collection
var ModificationsColl *mongo.Collection
var AnomaliesColl *mongo.Collection

type ConfigChange struct {
	InstanceId string `bson.D:"instanceId"`
	Operation  string `bson.D:"operation"`
	Path       string `bson.D:"path"`
}

func main() {

	if err := godotenv.Load("env_atlas"); err != nil {
		log.Fatal("Environment file not found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	MongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := MongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// Ping the primary
	if err := MongoClient.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	ModulesColl = MongoClient.Database("PhotonInsights").Collection("Modules")
	datagen.ModulesColl = ModulesColl
	MessagesColl = MongoClient.Database("PhotonInsights").Collection("Messages")
	datagen.MessagesColl = MessagesColl
	ModificationsColl = MongoClient.Database("PhotonInsights").Collection("Modifications")
	datagen.ModificationsColl = ModificationsColl
	AnomaliesColl = MongoClient.Database("PhotonInsights").Collection("Anomalies")

	startDate := time.Now()
	log.Printf("Started processing")
	for i := 0; i < 75000; i++ {
		datagen.GenerateTestMessage()
		if i%100 == 0 && i > 0 {
			log.Printf("Processed %d messages in %v seconds", i, time.Since(startDate).Seconds())
		}
	}
	endDate := time.Now()
	log.Printf("Completed processing in %v seconds", endDate.Sub(startDate).Seconds())

}
