package main

import (
	"errors"
	"fmt"
	"gbGATEWAY/gbp"
	"gbGATEWAY/mongoAction"
	"gbGATEWAY/redisAction"
	"gbGATEWAY/rmq"
	"gbGATEWAY/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

type clientDetail struct {
	con *websocket.Conn
	UId string
}

type MAIN struct {
	RedisDB *redisAction.Redis
	MongoDB *mongoAction.Mongo
	RMQ     *rmq.RMQ
	lock    *sync.RWMutex
	Name    string
	Epoll   *epoll
}

var (
	epoller     *epoll
	recvChannel                            = make(chan []byte)
	conChannel                             = make(chan websocket.Conn)
	clients     map[string]*clientDetail   = make(map[string]*clientDetail)
	connMap     map[*websocket.Conn]string = make(map[*websocket.Conn]string)
)

type epoll struct {
	fd          int
	connections map[int]websocket.Conn
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]websocket.Conn),
	}, nil
}

func (m *MAIN) Add(conn *websocket.Conn) error {

	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Println("[AddReadError1] : ", err.Error())
	}
	var trans gbp.Transport
	err = proto.Unmarshal(data, &trans)
	if err != nil {
		log.Println("[ProtoUNMError1] : ", err.Error())
		conn.Close()
		return err
	}
	if trans.Tp == 1 {
		smk, err := m.MongoDB.GetMainKey(trans.Id)
		if err != nil {
			log.Println("[NoKeyFoundError] : ", err.Error())
			conn.Close()
			return err
		}
		pT, err := utils.AesDecryption(utils.Decode(smk), trans.Msg)
		if err != nil {
			log.Println("[AESDECRERRROR1] : ", err.Error())
			conn.Close()
			return err
		}
		var nT gbp.ClientName
		err = proto.Unmarshal(pT, &nT)
		if err != nil {
			log.Println("[ProtoUNMError2] : ", err.Error())
			conn.Close()
			return err
		}

		var new clientDetail
		new.con = conn
		new.UId = nT.UId

		connMap[conn] = nT.MId

		clients[nT.MId] = &new
		err = m.RMQ.Produce(data)
		if err != nil {
			log.Println(err.Error())
			conn.Close()
			return err
		}
		// Extract file descriptor associated with the connection
		fd := websocketFD(conn)
		err = unix.EpollCtl(m.Epoll.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
		if err != nil {
			return err
		}
		m.Epoll.lock.Lock()
		defer m.Epoll.lock.Unlock()
		m.Epoll.connections[fd] = *conn
		m.RedisDB.Client.Set(nT.MId, m.Name, 0)
		fmt.Println("new client connected")
		return nil
	}
	return errors.New("[EADDERROR] : bad client detail")
}

func (e *epoll) Remove(conn websocket.Conn) error {
	fd := websocketFD(&conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}

	return nil
}

func (e *epoll) Wait() ([]*websocket.Conn, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []*websocket.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, &conn)
	}
	return connections, nil
}

func websocketFD(conn *websocket.Conn) int {
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func (m *MAIN) wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if err := m.Add(conn); err != nil {
		log.Println("Failed to add connection: ", err)
		conn.Close()
	}
}

func (m *MAIN) conClose() {
	for conn := range conChannel {
		fmt.Println("Removing...")
		epoller.Remove(conn)
		mid := connMap[&conn]
		m.RedisDB.Client.Del(mid)
		delete(clients, mid)
		delete(connMap, &conn)
		fmt.Println("Removed!")
	}
}

func (m *MAIN) dataProcess() {
	var err error
	for msg := range recvChannel {
		err = m.RMQ.Produce(msg)
		if err != nil {
			log.Println("[rmqERROR] : ", err.Error())
		}
		fmt.Println("job created")
	}
}

// Run in goroutine
func (m *MAIN) sender() {
	// TODO 1: recv job using rpc from engine
	// if tp = 1, then it is a send job, then consume the data from the rabbitmq channel of id that is mentioned in the job
	// 	then data from rmq and write to the specified client. And if in the mean time the client get disconnected from gateway
	// then persiste the data into mongodb.
	var sD gbp.SendNotify
	for msg := range m.RMQ.Msgs {
		fmt.Println("new sender")
		err := proto.Unmarshal(msg.Body, &sD)
		if err != nil {
			log.Println(err.Error())
		}
		fmt.Println("sD.TMid: ", sD.TMid)
		cD := clients[sD.TMid]
		fmt.Println("cD: ", cD)
		if cD != nil {
			err = cD.con.WriteMessage(2, sD.Data)
			if err != nil {
				log.Println("[WSWriteError] : ", err.Error())
				continue
			}
			fmt.Println("job prosseced")
		}
	}
}

func Start() {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if _, msg, err := conn.ReadMessage(); err != nil {
				conChannel <- *conn
				fmt.Println("[ERROR] -->> ", err.Error())
			} else {
				recvChannel <- msg
			}
		}
	}
}

func showError(heading string, err error) {
	red := color.New(color.FgRed).PrintfFunc()
	white := color.New(color.FgWhite).PrintfFunc()
	red(heading + " : ")
	white(err.Error())
	fmt.Println("")
}

func main() {
	// LOADING ENVIRONMENT VARIABLES
	godotenv.Load()

	mongoIP, found := os.LookupEnv("MONGO_LOC_IP")
	if !found {
		panic("environment variable missing, MONGO_LOC_IP")
	}

	rabbitIP, found := os.LookupEnv("RABBITMQ_LOC_IP")
	if !found {
		panic("environment variable missing, RABBITMQ_LOC_IP")
	}

	redisIP, found := os.LookupEnv("REDIS_LOC_IP")
	if !found {
		panic("environment variable missing, REDIS_LOC_IP")
	}

	rabbitUsername, found := os.LookupEnv("RABBITMQ_USERNAME")
	if !found {
		panic("environment variable missing, RABBITMQ_USERNAME")
	}

	rabbitPassword, found := os.LookupEnv("RABBITMQ_PASSWORD")
	if !found {
		panic("environment variable missing, RABBITMQ_PASSWORD")
	}

	mongoUsername, found := os.LookupEnv("MONGO_USERNAME")
	if !found {
		panic("environment variable missing, MONGO_USERNAME")
	}

	mongoPassword, found := os.LookupEnv("MONGO_PASSWORD")
	if !found {
		panic("environment variable missing, MONGO_PASSWORD")
	}

	var name string = "GT" + utils.Encode(utils.GenerateAesKey(10))

	// Increase resources limitations
	go func() {
		var rLimit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			panic(err)
		}
		rLimit.Cur = rLimit.Max
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			panic(err)
		}
	}()

	// // Enable pprof hooks
	// go func() {
	// 	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
	// 		log.Fatalf("pprof failed: %v", err)
	// 	}
	// }()

	// initlising MAIN struct
	var m MAIN
	m.lock = &sync.RWMutex{}
	clients = make(map[string]*clientDetail)
	m.Name = name

	// setting up mongodb
	var mongoDB mongoAction.Mongo
	mongoDB.Init(mongoIP, mongoUsername, mongoPassword)

	// setting up redisdb
	var redisDB redisAction.Redis
	redisDB.Init(redisIP)

	// Setting up rabbitmq
	var RMQ rmq.RMQ
	RMQ.Init(rabbitIP, rabbitUsername, rabbitPassword, name)

	// initilising all the fields in MAIN struct
	RMQ.RedisDB = &redisDB
	m.MongoDB = &mongoDB
	m.RedisDB = &redisDB
	m.RMQ = &RMQ

	// starting concurrent workers
	go m.sender()
	go m.conClose()
	go m.dataProcess()
	go Start()

	// Start epoll
	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}
	m.Epoll = epoller

	// starting websocket handler
	http.HandleFunc("/", m.wsHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
