package handler

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"gbGATEWAY/utils"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
)

type CacheHandler struct {
	RedisClient *redis.Client
	Logger      *utils.Logger
}

// func (cache *CacheHandler) GetRandomEngineName() (string, error) {
// 	ress := cache.RedisClient.LRange("engines", 0, -1)
// 	engines, err := ress.Result()
// 	if err != nil || len(engines) == 0 {
// 		return "", errors.New("no engine found")
// 	}

// 	rand.Seed(time.Now().UnixNano())
// 	// generate random number and print on console
// 	random := rand.Intn(len(engines))
// 	return engines[random], nil
// }

func (cache *CacheHandler) GetRandomEngineName() (string, error) {
	ress := cache.RedisClient.SInter("engines")
	engines, err := ress.Result()
	if err != nil || len(engines) == 0 {
		return "", errors.New("no engine found")
	}

	rand.Seed(time.Now().UnixNano())
	// generate random number and print on console
	random := rand.Intn(len(engines))
	return engines[random], nil
}

func (cache *CacheHandler) RegisterNode(nodeName string) error {
	result := cache.RedisClient.SAdd("gateways", nodeName)
	return result.Err()
}

func (cache *CacheHandler) RemoveNode(nodeName string) error {
	fmt.Println("Called using defer")
	result := cache.RedisClient.SRem("gateways", nodeName)
	return result.Err()
}

func (cache *CacheHandler) SetUserConnectNode(uuid string, nodeName string) error {
	sha := sha1.New()
	_, err := sha.Write([]byte(uuid))
	if err != nil {
		return err
	}

	hash := sha.Sum(nil)
	b64Uuid := utils.Encode(hash)
	res := cache.RedisClient.Set("CD_"+b64Uuid, nodeName, 0)
	return res.Err()
}

func (cache *CacheHandler) RemoveUserConnectNode(uuid string) error {
	sha := sha1.New()
	_, err := sha.Write([]byte(uuid))
	if err != nil {
		return err
	}

	hash := sha.Sum(nil)
	b64Uuid := utils.Encode(hash)
	result := cache.RedisClient.Del(b64Uuid)
	return result.Err()
}
