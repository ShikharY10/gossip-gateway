package config

import "github.com/go-redis/redis"

// Connect to Redis DataBase and return {*redis.Client} if successfully connected to redis server
func ConnectRedis(env *ENV) (*redis.Client, error) {
	options := redis.Options{
		Addr:     env.REDIS_HOST + ":" + env.REDIS_PORT,
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
