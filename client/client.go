package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Errorf("Connection lost: %v\n", err)
}

func registerDevice(name string, id string, uname string, pass string, broker string, port int) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(fmt.Sprintf("%s/%s", name, id))
	opts.SetUsername(uname)
	opts.SetPassword(pass)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	return client
}

func publishHeartbeat(client mqtt.Client) {
	for {
		message := strconv.FormatInt(time.Now().UnixNano(), 10)
		token := client.Publish("heartbeat", 0, false, message)
		token.Wait()
		time.Sleep(time.Second * 30)
	}
}

func main() {
	var port = 1883
	var id string

	var broker = flag.String("host", "localhost", "The hostname of the MQTT broker")
	var name = flag.String("name", "", "The name with which the device will be identified")
	var reRegister = flag.Bool("r", false, "Re-register the device with a new name and Id")
	var uname = flag.String("username", "ben", "Username for the MQTT connection")
	var pass = flag.String("password", "123", "Password for the MQTT connection")
	flag.Parse()

	if *reRegister {
		name, id = handle_reRegistration(*name)
	} else {
		if _, err := os.Stat("device.info"); errors.Is(err, os.ErrNotExist) {
			name, id = handleRegistration(*name)
		} else {
			name, id = handleGetDeviceInfo()
		}
	}
	client := registerDevice(*name, id, *uname, *pass, *broker, port)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	token := client.Publish("register", 0, false, fmt.Sprintf("%s,%s", *name, id))
	token.Wait()
	publishHeartbeat(client)
}

func SetWritable(filepath string) error {
	err := os.Chmod(filepath, 0777)
	return err
}

func SetReadOnly(filepath string) error {
	err := os.Chmod(filepath, 0444)
	return err
}

func handleGetDeviceInfo() (*string, string) {
	var content []byte
	var err error
	content, err = ioutil.ReadFile("device.info")
	if err != nil {
		log.Fatal(err)
	}
	res := strings.Split(string(content), ",")
	return &res[1], res[0]
}

func handleRegistration(name string) (*string, string) {
	var infoFile *os.File
	var err error
	infoFile, err = os.OpenFile("device.info", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer infoFile.Close()
	id := []byte(uuid.New().String())
	if len(name) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
		name = hostname
	}
	infoFile.WriteString(string(id) + "," + name)
	err = SetReadOnly("device.info")
	if err != nil {
		log.Fatal(err)
	}
	return &name, string(id)
}

func handle_reRegistration(name string) (*string, string) {
	var err = SetWritable("device.info")
	if err != nil {
		log.Fatal(err)
	}
	return handleRegistration(name)
}
