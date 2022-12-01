package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

var devices = make(map[string][]string)

var heartbeatMsgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}
var registerMsgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	info := strings.Split(string(msg.Payload()), ",")
	var val [3]string
	val[0] = info[0]
	registeredOn := strconv.FormatInt(time.Now().UnixNano(), 10)
	val[1] = registeredOn
	val[2] = registeredOn
	devices[info[1]] = val[:]
	log.Printf("%+v", devices)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Fatal("Connection lost: %v", err)
}

func sub(client mqtt.Client) {
	token := client.Subscribe("heartbeat", 1, heartbeatMsgHandler)
	token.Wait()
	log.Println("Subscribed to topic: heartbeat")
	token = client.Subscribe("register", 1, registerMsgHandler)
	token.Wait()
	log.Println("Subscribed to topic: register")
}

func main() {
	var broker = "localhost"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("benmeehan111/server")
	opts.SetUsername("ben")
	opts.SetPassword("123")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	sub(client)
	select {}
}
