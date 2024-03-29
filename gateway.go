package main

import (
	"fmt"
	"gbGATEWAY/admin"
	"gbGATEWAY/config"
	"gbGATEWAY/controllers"
	"gbGATEWAY/epoll"
	"gbGATEWAY/handler"
	"gbGATEWAY/middleware"
	"gbGATEWAY/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func removeUsers(channel chan string, cache *handler.CacheHandler) {
	for userId := range channel {
		cache.RemoveUserConnectNode(userId)
	}
}

func readDataFromClient(channel chan []byte, handler *handler.Handler) {

	for job := range channel {
		engine, err := handler.Cache.GetRandomEngineName()

		if err != nil {
			// ToDo: Log this error
			fmt.Println("ERROR: ", err.Error())
			continue
		}
		err = handler.Queue.Produce(engine, job)
		if err != nil {
			log.Println("[PRODUCE ERROR] : ", err.Error())
		}
	}
}

func main() {
	// LOADING ENVIRONMENT VARIABLES
	ENV := config.LoadENV()

	logger, err := admin.InitializeLogger(ENV, "gateway")
	if err != nil {
		log.Fatal(err)
	}

	db, err := config.ConnectToDBs(ENV)
	if err != nil {
		log.Fatal(err)
	}

	queue, err := config.ConnectToQueue(ENV)
	if err != nil {
		log.Fatal(err)
	}

	Epoll, err := epoll.InitiatEpoll(logger)
	if err != nil {
		log.Fatal(err)
	}

	cache := &handler.CacheHandler{
		RedisClient: db.RedisDB,
		Logger:      logger,
	}

	handle := handler.Handler{
		Database: &handler.DataBaseHandler{
			Mongo:  *db.MongoDB,
			Logger: logger,
		},
		Cache: cache,
		Queue: &handler.QueueHandler{
			Queue:   *queue,
			Clients: Epoll.Clients,
			Logger:  logger,
		},
		Env: ENV,
	}

	middleWare := middleware.CreateMiddleware([]byte(ENV.JWT_ACCESS_TOKEN_SECRET_KEY), cache)

	controller := controllers.Controller{
		Epoll:      Epoll,
		Handler:    &handle,
		Middleware: middleWare,
	}

	go handle.Queue.Consume()
	go readDataFromClient(Epoll.DataPipeline, &handle)
	go removeUsers(Epoll.ClosePipeline, handle.Cache)
	go Epoll.StartEpollMonitoring()

	router := gin.New()
	routes.WebsocketRoute(router, controller)

	admin.ShowSucces("Starting Gateway at PORT => ["+ENV.GatewayPort+"]", true)
	router.Run(":" + ENV.GatewayPort)
}
