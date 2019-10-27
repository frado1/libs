package mqtthelper

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	flags "github.com/jessevdk/go-flags"
	yaml "gopkg.in/yaml.v2"
)

type OnConnectHandler func(c mqtt.Client)
type MessageChannel chan mqtt.Message

type opts struct {
	ConfigFile string `short:"c" long:"config" default:"config.yaml" description:"Path to config file to use"`
}

var delays = make(map[string]*time.Timer)
var delayMutex = &sync.Mutex{}

func ParseConfigOption(c interface{}) error {
	opts := opts{}
	if _, err := flags.Parse(&opts); err != nil {
		return err
	}

	if err := LoadConfig(opts.ConfigFile, c); err != nil {
		return err
	}

	return nil
}

func LoadConfig(path string, c interface{}) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(f, c); err != nil {
		return err
	}

	return nil
}

func NewClientLogin(uri string, user string, password string, h OnConnectHandler) (mqtt.Client, error) {
        co := getClientOptions(uri, h, true)
        co.SetUsername(user) 
        co.SetPassword(password)
        c := mqtt.NewClient(co)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func NewClient(uri string, h OnConnectHandler) (mqtt.Client, error) {
	c := mqtt.NewClient(getClientOptions(uri, h, false))
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func NewClientParallel(uri string, h OnConnectHandler) (mqtt.Client, error) {
	c := mqtt.NewClient(getClientOptions(uri, h, true))
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func NewClientParallelLogin(uri string, user string, password string, h OnConnectHandler) (mqtt.Client, error) {
        co := getClientOptions(uri, h, true)
        co.SetUsername(user) 
        co.SetPassword(password)
        c := mqtt.NewClient(co)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func NewMessageChannel() MessageChannel {
	return make(MessageChannel)
}

func Subscribe(c mqtt.Client, topic string, ch MessageChannel) error {
	h := func(c mqtt.Client, msg mqtt.Message) {
		ch <- msg
	}

	return SubscribeHandler(c, topic, h)
}

func SubscribeHandler(c mqtt.Client, topic string, handler mqtt.MessageHandler) error {
	h := func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message '%s' through topic %s (retained: %s)", msg.Payload(), msg.Topic(), strconv.FormatBool(msg.Retained()))
		handler(c, msg)
	}

	if token := c.Subscribe(topic, 0, h); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func PublishMessage(c mqtt.Client, topic string, qos byte, retained bool, payload string) bool {
	token := c.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("Could not publish message '%s' to topic %s: %s", payload, topic, token.Error())
		return false
	}
	log.Printf("Published message '%s' to topic %s (retained: %s)", payload, topic, strconv.FormatBool(retained))
	return true
}

func PublishCustomMessage(c mqtt.Client, topic string, qos byte, retained bool, payload interface{}) bool {
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error while marshalling message payload for topic %s: %s", topic, err)
		return false
	}

	return PublishMessage(c, topic, qos, retained, string(p))
}

func DelayMessage(c mqtt.Client, id string, topic string, qos byte, retained bool, payload string, delay time.Duration) {
	if _, ok := delays[id]; ok {
		return
	}

	f := func() {
		PublishMessage(c, topic, qos, retained, payload)
		CancelDelayedMessage(id)
	}

	log.Printf("Delay message with id %s on topic %s", id, topic)
	delayMutex.Lock()
	delays[id] = time.AfterFunc(delay, f)
	delayMutex.Unlock()

	return
}

func CancelDelayedMessage(id string) {
	newDelays := make(map[string]*time.Timer)

	delayMutex.Lock()
	for delayId, t := range delays {
		if delayId == id {
			t.Stop()
		} else {
			newDelays[id] = t
		}
	}
	delays = newDelays
	delayMutex.Unlock()
}

func HandleError(err error, msgPrefix string) {
	if err != nil {
		log.Printf("%s: %s", msgPrefix, err)
	}
}

func HandleFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getClientOptions(uri string, h OnConnectHandler, parallel bool) *mqtt.ClientOptions {
	ops := mqtt.NewClientOptions().AddBroker(uri)

	if parallel {
		ops.SetOrderMatters(false)
	}

	ops.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("Connection to MQTT lost: %s", err)
	})

	ops.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("Connected to MQTT at %s", uri)
		h(c)
	})

	return ops
}
