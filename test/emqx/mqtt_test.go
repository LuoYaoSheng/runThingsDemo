package emqx

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"

	"strconv"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MessagePubHandler mqtt.MessageHandler = func(mqttClient mqtt.Client, msg mqtt.Message) {
	fmt.Printf("+++++++++++Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func TestMqtt(t *testing.T) {

	mqttCfg := &config.MqttConf{
		Broker:   "ws://127.0.0.1:8083/mqtt",
		Username: "",
		Password: "",
		Topic:    "log/#",
		Qos:      0,
	}

	// 开启订阅模式
	err := service.MqttSubscribe(mqttCfg.Broker, mqttCfg.Username, mqttCfg.Password, mqttCfg.Topic, mqttCfg.Qos, MessagePubHandler)
	if err != nil {
		log.Panicln(err)
		return
	}

	// 发送一条 日志
	m := make(map[string]interface{})
	m["name"] = "智能井盖"
	m["location"] = "智慧展厅"
	m["time"] = time.Now()

	mqLog := model.Eq2MqLog{
		Sn:      "eq001",
		Product: "p1001",
		Status:  config.EqStatusAlarm,
		Content: m,
		Title:   "发生倾斜",
	}

	topic := "log/" + mqLog.Product + "/" + mqLog.Sn + "/" + strconv.FormatInt(mqLog.Status, 10)
	content, _ := json.Marshal(mqLog)

	err2 := service.MqttSend(topic, content, mqttCfg.Qos)
	if err2 != nil {
		log.Panicln(err2)
		return
	}
}
