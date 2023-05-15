package admin

import (
	"encoding/json"
	"fmt"
	"gbGATEWAY/config"
	"gbGATEWAY/models"
	"gbGATEWAY/utils"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type Logger struct {
	conn        *websocket.Conn
	ServiceName string
	serviceType string
}

func InitializeLogger(env *config.ENV, serviceType string) (*Logger, error) {
	url := "ws://" + env.LogServerHost + ":" + env.LogServerPort + "/connect?name=" + env.GateWayName + "&type=" + serviceType + "&port="
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	ShowSucces("connected to log server", false)
	logger := &Logger{
		conn:        conn,
		ServiceName: env.GateWayName,
		serviceType: serviceType,
	}
	return logger, nil
}

func (l *Logger) RegisterNewLog(log models.Log) {
	logBytes, _ := json.Marshal(&log)

	var packet models.Packet
	packet.NodeName = l.ServiceName
	packet.Type = "log"
	packet.Message = utils.Encode(logBytes)

	packetBytes, _ := json.Marshal(&packet)
	l.conn.WriteMessage(2, packetBytes)
}

func (l *Logger) LogError(err error) {
	_, fileLocation, line, _ := runtime.Caller(1)
	log := models.Log{
		TimeStamp:   time.Now().String(),
		ServiceType: l.serviceType,
		Type:        "ERROR",
		FileName:    fileLocation,
		LineNumber:  line - 2,
		Message:     err.Error(),
	}
	l.RegisterNewLog(log)
}

func ShowError(heading string, err error) {
	red := color.New(color.FgRed).PrintfFunc()
	white := color.New(color.FgWhite).PrintfFunc()
	red(heading + " : ")
	white(err.Error())
	fmt.Println("")
}

func ShowSucces(message string, major bool) {
	var messagePrinter func(format string, a ...interface{})
	if major {
		messagePrinter = color.New(color.FgBlue, color.Bold, color.Underline).PrintfFunc()
	} else {
		messagePrinter = color.New(color.FgWhite).PrintfFunc()
	}
	head := color.New(color.FgGreen).PrintfFunc()
	head("[SUCCESS] -> ")
	messagePrinter(message)
	fmt.Println("")
}
