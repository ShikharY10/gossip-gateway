package handler

import "gbGATEWAY/config"

type Handler struct {
	Database *DataBaseHandler
	Cache    *CacheHandler
	Queue    *QueueHandler
	Env      *config.ENV
}
