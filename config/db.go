package config

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBase struct {
	MongoDB *MongoDB
	RedisDB *redis.Client
}

type MongoDB struct {
	Users *mongo.Collection
}

func ConnectToDBs(env *ENV) (*DataBase, error) {
	var db DataBase
	mongoClient, err := db.mongoDB(env)
	if err != nil {
		return nil, err
	}
	redisClient, err := db.redisDB(env)
	if err != nil {
		return nil, err
	}
	return &DataBase{
		MongoDB: mongoClient,
		RedisDB: redisClient,
	}, nil
}

func (db *DataBase) mongoDB(env *ENV) (*MongoDB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var mongoClient *mongo.Client
	var err error

	if env.MongoDBConnectionMethod == "manual" {
		credential := options.Credential{
			Username: env.MongoDBUsername,
			Password: env.MongoDBPassword,
		}

		clientOptions := options.Client().ApplyURI("mongodb://" + env.MongoDBHost + ":" + env.MongoDBPort).SetAuth(credential)
		mongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			defer cancel()
			return nil, err
		}
	} else if env.MongoDBConnectionMethod == "auto" {
		serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
		clientOptions := options.Client().ApplyURI(env.MongoDBConnectionString).SetServerAPIOptions(serverAPIOptions)
		mongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			defer cancel()
			return nil, err
		}
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		defer cancel()
		return nil, err
	}

	var mongo MongoDB
	storage := mongoClient.Database("storage")
	mongo.Users = storage.Collection("users")

	defer cancel()
	return &mongo, nil
}

func (db *DataBase) redisDB(env *ENV) (*redis.Client, error) {
	redisIP := "127.0.0.1"
	options := redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       0,
	}
	client := redis.NewClient(&options)
	ping := client.Ping()
	if ping.Err() != nil {
		return nil, ping.Err()
	}
	return client, nil
}
