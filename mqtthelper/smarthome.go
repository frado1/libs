package mqtthelper

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SmartHomeBroker struct {
	c             mqtt.Client
	TopLevelTopic string
}

type SmartHomeOnConnectHandler func(b SmartHomeBroker)

func NewSmartHomeBroker(uri string, topLevelTopic string, h SmartHomeOnConnectHandler) SmartHomeBroker {
	broker := SmartHomeBroker{
		TopLevelTopic: topLevelTopic,
	}

	ops := mqtt.NewClientOptions().AddBroker(uri)

	ops.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("Connection to MQTT lost: %s", err)
	})

	ops.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("Connected to MQTT at %s", uri)
		broker.setConnectionState("2")
		h(broker)
	})

	ops.SetWill(broker.connectedTopic(), "0", 0, true)

	broker.c = mqtt.NewClient(ops)

	return broker
}

func (b SmartHomeBroker) Connect() error {
	if token := b.c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (b SmartHomeBroker) Disconnect() {
	b.setConnectionState("0")
	b.c.Disconnect(100)
}

func (b SmartHomeBroker) Subscribe(topic string, ch MessageChannel) error {
	h := func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message '%s' through topic %s (retained: %s)", msg.Payload(), msg.Topic(), strconv.FormatBool(msg.Retained()))
		ch <- msg
	}

	if token := b.c.Subscribe(topic, 0, h); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (b SmartHomeBroker) PublishStatus(item string, payload string) bool {
	return b.publish(b.statusTopic(item), 0, true, payload)
}

func (b SmartHomeBroker) publish(topic string, qos byte, retained bool, payload string) bool {
	token := b.c.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("Could not publish message '%s' to topic %s: %s", payload, topic, token.Error())
		return false
	}
	log.Printf("Published message '%s' to topic %s (retained: %s)", payload, topic, strconv.FormatBool(retained))
	return true
}

func (b SmartHomeBroker) setConnectionState(state string) {
	PublishMessage(b.c, b.connectedTopic(), 0, true, state)
}

func (b SmartHomeBroker) connectedTopic() string {
	return b.TopLevelTopic + "/connected"
}

func (b SmartHomeBroker) statusTopic(item string) string {
	return b.TopLevelTopic + "/status/" + item
}
