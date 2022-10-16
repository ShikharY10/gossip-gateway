package redisAction

import (
	"fmt"
	"gbGATEWAY/utils"
	"log"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/go-redis/redis"
)

type Redis struct {
	Client *redis.Client
}

func (r *Redis) Init(redisIP string) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       0,
	})
	s := client.Ping()
	fmt.Println(s.String())

	client.Set("name", "Shikhar Yadav", 0)
	res := client.Get("name")
	fmt.Println(res.Result())
	r.Client = client
	color.Green("Redis connected!")
	// color.GreenString
	// fmt.Println("Redis client connected!")
}

func (r *Redis) GetGatewayName() []string {
	ress := r.Client.LRange("engines", 0, -1)
	gateways, err := ress.Result()
	if err != nil {
		log.Println(err.Error())
	}
	return gateways
}

func (r *Redis) GetEngineName() []string {
	ress := r.Client.LRange("engines", 0, -1)
	engines, err := ress.Result()
	if err != nil {
		log.Println(err.Error())
	}
	return engines
}

func (r *Redis) SetGatewayName(name string) error {
	res := r.Client.RPush("gateways", name)
	return res.Err()
}

func (r *Redis) SetUserData(id int, data map[string]interface{}) {
	key := strconv.Itoa(id) + "data"
	status := r.Client.HMSet(key, data)
	s, e := status.Result()
	if e != nil {
		panic(e)
	}
	fmt.Println("s: ", s)
}

func (r *Redis) GetUserIsOnline(id int) int {
	key := strconv.Itoa(id) + "data"
	value := r.Client.HMGet(key, "IsOnline")
	val, err := value.Result()
	if err != nil {
		panic(err)
	}

	v := val[0].(string)
	i, e := strconv.Atoi(v)
	if e != nil {
		panic(e)
	}
	return i
}

func (r *Redis) GetUserLastSeen(id int) string {
	key := strconv.Itoa(id) + "data"
	value := r.Client.HMGet(key, "LastSeen")
	val, err := value.Result()
	if err != nil {
		panic(err)
	}

	v := val[0].(string)
	return v
}

func (r *Redis) GetUserServerName(id int) int {
	key := strconv.Itoa(id) + "data"
	value := r.Client.HMGet(key, "servername")
	val, err := value.Result()
	if err != nil {
		panic(err)
	}

	v := val[0].(string)
	i, e := strconv.Atoi(v)
	if e != nil {
		panic(e)
	}
	return i
}

// func (r *Redis) GetClientData(id int) utils.UserData {
// 	key := strconv.Itoa(id) + "data"
// 	value := r.Client.HMGet(key, "IsOnline", "LastSeen", "servername")
// 	in := value.Val()
// 	var ud utils.UserData
// 	ud.LastSeen = in[1].(string)
// 	i, e := strconv.Atoi(in[0].(string))
// 	if e != nil {
// 		panic(e)
// 	}
// 	ud.IsOnline = i

// 	ir, er := strconv.Atoi(in[0].(string))
// 	if er != nil {
// 		panic(er)
// 	}
// 	ud.IsOnline = ir

// 	return ud
// }

func (r *Redis) GetUserMsg(id int) (string, error) {
	key := strconv.Itoa(id)
	s := r.Client.LPop(key)
	str, err := s.Result()
	if err != nil {
		panic(err)
	}
	return str, nil
}

func (r *Redis) SetUserMsg(id int, msg string) error {
	key := strconv.Itoa(id)
	s := r.Client.RPush(key, msg)
	_, e := s.Result()
	if e != nil {
		return e
	}
	return nil
}

func (r *Redis) RegisterOTP() (string, string) {
	id64 := utils.GenerateRandomId()
	otp := utils.GenerateOTP(6)
	r.Client.Set(id64, otp, time.Duration(5*time.Minute))
	return id64, otp
}

func (r *Redis) VarifyOTP(id string, otp string) bool {
	res := r.Client.Get(id)
	_otp := res.Val()
	if otp == _otp {
		return true
	} else {
		return false
	}
}
