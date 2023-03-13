package utils

import (
	"encoding/json"
	"fmt"
	"gbGATEWAY/config"
	"gbGATEWAY/schema"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type Logger struct {
	conn *websocket.Conn
	name string
}

func InitializeLogger(env *config.ENV, serviceType string) (*Logger, error) {
	url := "ws://" + env.LogServerHost + ":" + env.LogServerPort + "/connect?name=" + env.GateWayName + "&type=" + serviceType + "&port=" + env.GatewayPort
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	ShowSucces("connected to log server", false)
	logger := &Logger{
		conn: conn,
		name: env.GateWayName,
	}

	logger.testLogger()
	return logger, nil
}

func (l *Logger) testLogger() {
	log := schema.Log{
		TimeStamp:   time.Now().String(),
		ServiceType: "gateway",
		Type:        "Test",
		FileName:    "logger.go",
		LineNumber:  31,
		Message:     "Testing Logger from gateway",
	}
	l.RegisterNewLog(log)
}

func (l *Logger) RegisterNewLog(log schema.Log) {
	logBytes, _ := json.Marshal(&log)

	var packet schema.Packet
	packet.NodeName = l.name
	packet.Type = "log"
	packet.Message = Encode(logBytes)

	packetBytes, _ := json.Marshal(&packet)
	l.conn.WriteMessage(2, packetBytes)
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
