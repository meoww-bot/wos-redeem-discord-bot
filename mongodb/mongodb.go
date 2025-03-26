package mongodb

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client = MongoConn()

func MongoConn() *mongo.Client {

	var MongoURI = os.Getenv("MongoURI")

	if MongoURI == "" {
		log.Println("MongoURI is not set in the environment variables")
		os.Exit(-1)
	}

	clientOptions := options.Client().ApplyURI(MongoURI) //.SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("Failed to connect to MongoDB ,err:" + err.Error())
		os.Exit(1)
	}

	log.Println("Connected to MongoDB!")
	return client
}

func getContext() (ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	return
}
