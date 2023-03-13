package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"

	"github.com/joho/godotenv"
)

type ENV struct {
	MongoDBConnectionMethod string // manual
	MongoDBPort             string // 27017
	MongoDBHost             string // 127.0.0.1
	MongoDBUsername         string // rootuser
	MongoDBPassword         string // rootpass
	MongoDBConnectionString string // mongodb connection string will be used when MongoDBConnectionMethod is set to auto
	RedisHost               string // 127.0.0.1
	RedisPort               string // 6379
	RabbitMQHost            string // 127.0.0.1
	RabbitMQPort            string // 5672
	RabbitMQUsername        string // guest
	RabbitMQPassword        string // guest
	GatewayPort             string // 6001
	GateWayName             string // GT____
	GateWayMode             string // debug
	JWTSecret               string // abcdefghijklmnopqrstuvwxyz
	LogServerHost           string // 127.0.0.1
	LogServerPort           string // 6002
}

// Generate fixed size byte array
func generateRandomId(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func encode(data []byte) string {
	hb := base64.StdEncoding.EncodeToString([]byte(data))
	return hb
}

func LoadENV() *ENV {
	godotenv.Load()
	var env ENV

	var value string
	var found bool

	value, found = os.LookupEnv("MONGODB_CONNECTION_METHOD")
	if found {
		env.MongoDBConnectionMethod = value
	} else {
		env.MongoDBConnectionMethod = "manual"
	}

	value, found = os.LookupEnv("MONGODB_PORT")
	if found {
		env.MongoDBPort = value
	} else {
		env.MongoDBPort = "27017"
	}

	value, found = os.LookupEnv("MONGODB_HOST")
	if found {
		env.MongoDBHost = value
	} else {
		env.MongoDBHost = "127.0.0.1"
	}

	value, found = os.LookupEnv("MONGODB_USERNAME")
	if found {
		env.MongoDBUsername = value
	} else {
		env.MongoDBUsername = "rootuser"
	}

	value, found = os.LookupEnv("MONGODB_PASSWORD")
	if found {
		env.MongoDBPassword = value
	} else {
		env.MongoDBPassword = "rootpass"
	}

	value, found = os.LookupEnv("MONGODB_CONNECTION_STRING")
	if found {
		env.MongoDBConnectionString = value
	} else {
		env.MongoDBConnectionString = ""
	}

	value, found = os.LookupEnv("REDIS_HOST")
	if found {
		env.RedisHost = value
	} else {
		env.RedisHost = "127.0.0.1"
	}

	value, found = os.LookupEnv("REDIS_PORT")
	if found {
		env.RedisPort = value
	} else {
		env.RedisPort = "6379"
	}

	value, found = os.LookupEnv("RabbitMQHost")
	if found {
		env.RabbitMQHost = value
	} else {
		env.RabbitMQHost = "127.0.0.1"
	}

	value, found = os.LookupEnv("RABBITMQ_PORT")
	if found {
		env.RabbitMQPort = value
	} else {
		env.RabbitMQPort = "5672"
	}

	value, found = os.LookupEnv("RABBITMQ_USERNAME")
	if found {
		env.RabbitMQUsername = value
	} else {
		env.RabbitMQUsername = "guest"
	}

	value, found = os.LookupEnv("RABBITMQ_PASSWORD")
	if found {
		env.RabbitMQPassword = value
	} else {
		env.RabbitMQPassword = "guest"
	}

	value, found = os.LookupEnv("GATEWAY_PORT")
	if found {
		env.GatewayPort = value
	} else {
		env.GatewayPort = "10222"
	}

	value, found = os.LookupEnv("GATEWAY_NAME")
	if found {
		env.GateWayName = value
	} else {
		env.GateWayName = "GT_" + encode(generateRandomId(10))
	}

	value, found = os.LookupEnv("GATEWAY_MODE")
	if found {
		env.GateWayMode = value
	} else {
		env.GateWayMode = "debug"
	}

	value, found = os.LookupEnv("JWTSECRET")
	if found {
		env.JWTSecret = value
	} else {
		env.JWTSecret = "abcdefghijklmnopqrstuvwxyz"
	}

	value, found = os.LookupEnv("LOG_SERVER_HOST")
	if found {
		env.LogServerHost = value
	} else {
		env.LogServerHost = "127.0.0.1"
	}

	value, found = os.LookupEnv("LOG_SERVER_PORT")
	if found {
		env.LogServerPort = value
	} else {
		env.LogServerPort = "10223"
	}

	return &env
}
