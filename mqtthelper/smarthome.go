package mqtthelper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SmartHomeBroker represents a broker
type SmartHomeBroker struct {
	mqttClient              mqtt.Client
	URI                     string
	TopLevelTopic           string
	OnConnectHandler        SmartHomeOnConnectHandler
	OnConnectionLostHandler SmartHomeOnConnectionLostHandler
	msgCh                   chan messageToHandle
}

// SmartHomeOnConnectHandler represents a callback when a connection to MQTT was established
type SmartHomeOnConnectHandler func(*SmartHomeBroker)

// SmartHomeOnConnectionLostHandler represents a callback when a connection to MQTT was lost
type SmartHomeOnConnectionLostHandler func(*SmartHomeBroker)

// SmartHomeMessageHandler represents a callback to handle received messages
type SmartHomeMessageHandler func(*SmartHomeBroker, mqtt.Message)

type messageToHandle struct {
	handler SmartHomeMessageHandler
	message mqtt.Message
}

// NewSmartHomeBroker creates a new SmartHomeBroker
func NewSmartHomeBroker(uri string, topLevelTopic string) *SmartHomeBroker {
	return &SmartHomeBroker{
		URI:           uri,
		TopLevelTopic: topLevelTopic,
		msgCh:         make(chan messageToHandle),
	}
}

// Connect tries to establish a connection to MQTT
func (b *SmartHomeBroker) Connect() error {
	if b.mqttClient == nil {
		b.mqttClient = mqtt.NewClient(b.getOptions())
	}

	if token := b.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Could not connect to MQTT at %s: %s", b.URI, token.Error())
	}

	return nil
}

// Disconnect closes the connection to MQTT
func (b *SmartHomeBroker) Disconnect() {
	if b.mqttClient == nil {
		return
	}
	b.publish(b.connectedTopic(), 0, true, "0")
	b.mqttClient.Disconnect(100)
}

// SetConnectionState sets the current state of the connection ("0": Disconnected from MQTT, "1": Connected to MQTT, but disconnected from hardware, "2": Fully operational)
func (b *SmartHomeBroker) SetConnectionState(connected bool) error {
	if b.mqttClient == nil {
		return fmt.Errorf("Not connected to MQTT, cannot set connection state to %s", strconv.FormatBool(connected))
	}
	if connected {
		b.publish(b.connectedTopic(), 0, true, "2")
	} else {
		b.publish(b.connectedTopic(), 0, true, "1")
	}

	return nil
}

// SubscribeAction registers a subscription to actions of the specified item
func (b *SmartHomeBroker) SubscribeAction(item string, h SmartHomeMessageHandler) error {
	return b.Subscribe(b.actionTopic(item), h)
}

// Subscribe registers a subscription to the specified topic
func (b *SmartHomeBroker) Subscribe(topic string, h SmartHomeMessageHandler) error {
	if b.mqttClient == nil {
		return fmt.Errorf("Not connected to MQTT, cannot subscribe to %s", topic)
	}

	f := func(mqttClient mqtt.Client, msg mqtt.Message) {
		b.msgCh <- messageToHandle{
			handler: h,
			message: msg,
		}
	}

	if token := b.mqttClient.Subscribe(topic, 0, f); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to subscribe to topic %s: %s", topic, token.Error())
	}
	return nil
}

// PublishSimpleStatus sends a simple status message for the specified item
func (b *SmartHomeBroker) PublishSimpleStatus(item string, payload string) error {
	if b.mqttClient == nil {
		return fmt.Errorf("Not connected to MQTT, cannot publish simple status for %s", item)
	}

	return b.publish(b.statusTopic(item), 0, true, payload)
}

// PublishStatus sends a status message for the specified item
func (b *SmartHomeBroker) PublishStatus(item string, payload interface{}) error {
	if b.mqttClient == nil {
		return fmt.Errorf("Not connected to MQTT, cannot publish simple status for %s", item)
	}

	p, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON for status of %s: %s", item, err)
	}

	return b.publish(b.statusTopic(item), 0, true, string(p))
}

// Run starts the main loop of the broker
func (b *SmartHomeBroker) Run() error {
	if err := b.Connect(); err != nil {
		return err
	}
	defer b.Disconnect()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	daemon.SdNotify(false, "READY=1")

	running := true
	for running {
		select {
		case msgToHandle := <-b.msgCh:
			msg := msgToHandle.message
			log.Printf("Received message '%s' through topic %s (retained: %s)", msg.Payload(), msg.Topic(), strconv.FormatBool(msg.Retained()))
			msgToHandle.handler(b, msg)
		case <-signalChannel:
			running = false
		}
	}

	return nil
}

func (b *SmartHomeBroker) getOptions() *mqtt.ClientOptions {
	ops := mqtt.NewClientOptions().AddBroker(b.URI)

	ops.SetConnectionLostHandler(func(mqttClient mqtt.Client, err error) {
		log.Printf("Connection to MQTT at %s lost: %s", b.URI, err)
		b.publish(b.connectedTopic(), 0, true, "0")
		if nil != b.OnConnectionLostHandler {
			b.OnConnectionLostHandler(b)
		}
	})

	ops.SetOnConnectHandler(func(mqttClient mqtt.Client) {
		log.Printf("Connected to MQTT at %s", b.URI)
		b.publish(b.connectedTopic(), 0, true, "1")
		if nil != b.OnConnectHandler {
			b.OnConnectHandler(b)
		}
	})

	ops.SetWill(b.connectedTopic(), "0", 0, true)

	return ops
}

func (b *SmartHomeBroker) publish(topic string, qos byte, retained bool, payload string) error {
	token := b.mqttClient.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("Could not publish message '%s' to topic %s: %s", payload, topic, token.Error())
	}
	log.Printf("Published message '%s' to topic %s (retained: %s)", payload, topic, strconv.FormatBool(retained))
	return nil
}

func (b *SmartHomeBroker) connectedTopic() string {
	return b.TopLevelTopic + "/connected"
}

func (b *SmartHomeBroker) actionTopic(item string) string {
	return b.TopLevelTopic + "/set/" + item
}

func (b *SmartHomeBroker) statusTopic(item string) string {
	return b.TopLevelTopic + "/status/" + item
}
