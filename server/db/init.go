// Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
// Mongo connection URI + database name are now read from the environment
// (MONGO_URI / MONGO_DB) so the server can run against a containerized Mongo.
package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"schej.it/server/logger"
)

var Client *mongo.Client
var Db *mongo.Database
var EventsCollection *mongo.Collection
var UsersCollection *mongo.Collection
var DailyUserLogCollection *mongo.Collection
var FriendRequestsCollection *mongo.Collection
var EventResponsesCollection *mongo.Collection
var AttendeesCollection *mongo.Collection

func Init() func() {
	// Read connection settings from the environment (with sensible local defaults)
	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		mongoUri = "mongodb://localhost"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "schej-it"
	}

	// Establish mongodb connection
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		logger.StdErr.Panicln(err)
	}

	// Define mongodb database + collections
	Db = Client.Database(dbName)
	EventsCollection = Db.Collection("events")
	UsersCollection = Db.Collection("users")
	DailyUserLogCollection = Db.Collection("dailyuserlogs")
	FriendRequestsCollection = Db.Collection("friendrequests")
	EventResponsesCollection = Db.Collection("eventResponses")
	AttendeesCollection = Db.Collection("attendees")

	// Return a function to close the connection
	return func() {
		Client.Disconnect(ctx)
	}
}

// MongoDB backup / restore commands

// Backup
// mongodump --uri="mongodb://localhost:27017" --db=schej-it

// Restore
// mongorestore --uri="mongodb://localhost:27017" --drop --db=schej-it ./dump
