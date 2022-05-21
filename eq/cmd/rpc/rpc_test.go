package rpc

import (
	"encoding/json"
	"log"
	model2 "run-things-demo/eq/model"

	"testing"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"
)

func recieveSimple(str string) {
	log.Println("---rpc_test: ", str)

	cmd := &model.Eq2MqCmd{}
	err := json.Unmarshal([]byte(str), &cmd)
	if err != nil {
		log.Panicln(err)
		return
	}

	topic := `th-calc/` + model2.ProductKey + `/` + cmd.Sn + `/cmd`
	service.MqttSend(topic, str, mqttCfg.Qos) // 直接透传 -- 可以的话，去掉 sn 也可以，减少传输内容，毕竟硬件空间比较小
}

var mqttCfg *config.MqttConf

func TestReceiveSimpleCmd(t *testing.T) {

	mqttCfg = &config.MqttConf{
		Broker:   "ws://127.0.0.1:8083/mqtt",
		Username: "",
		Password: "",
		Topic:    "",
		Qos:      0,
	}

	// 开启非订阅模式
	err := service.MqttSubscribe(mqttCfg.Broker, mqttCfg.Username, mqttCfg.Password, mqttCfg.Topic, mqttCfg.Qos, nil)
	if err != nil {
		log.Panicln(err)
		return
	}

	product := model2.ProductKey
	rabbitmq := service.NewRabbitMQSimple("runThings-cmd-"+product, "amqp://admin:admin@127.0.0.1:5672/")
	rabbitmq.ConsumeSimple(recieveSimple)
}
