# GOSSIP-GATEWAY

### A PART OF GOSSIP INTERNAL ARCHITECTURE...

A websocket gateway that is just used to provide real-time functionality.
It just take the data from user in real-time and forwarded it to gossip-engines.

### How to Deploy:

First of all we need to create .env file at the root of the project In which all the environment variables will be added.

Below are the required environment variables with there default values

```
    MONGO_LOC_IP=127.0.0.1
    MONGO_USERNAME=rootuser
    MONGO_PASSWORD=rootpass
    REDIS_LOC_IP=127.0.0.1
    RABBITMQ_LOC_IP=127.0.0.1
    RABBITMQ_USERNAME=guest
    RABBITMQ_PASSWORD=guest
```

It assume that we have running instances of MongoDB, Redis and RabbitMQ, and we have write the IP address of these services in .env file.

At last, run it by first building it according to your Operating System (It only support operating system that are based on unix)
or we can run it using command `go run gateway.go`
