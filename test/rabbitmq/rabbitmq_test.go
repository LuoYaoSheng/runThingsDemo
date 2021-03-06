package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"
)

var rabbitMqCfg *config.RabbitMqConf

func createCfg() {
	rabbitMqCfg = &config.RabbitMqConf{
		Mqurl:     "amqp://admin:admin@127.0.0.1:5672",
		QueueName: "runThings",
		Exchange:  "runThings",
		Key:       "runThings",
	}
}

func recieveSub(str string) {
	var mqLog model.Eq2MqLog
	err := json.Unmarshal([]byte(str), &mqLog)
	if err != nil {
		log.Panicln(err)
		return
	}
	fmt.Println("---接收到数据: ", mqLog.Sn, mqLog.Status, mqLog.Title)
}

// 先启动接收 --- 订阅模式
func TestRabbitMQReceiveSub(t *testing.T) {
	createCfg()

	rabbitmq := service.NewRabbitMQPubSub(rabbitMqCfg.Exchange, rabbitMqCfg.Mqurl)
	rabbitmq.RecieveSub(recieveSub)
}

// 再启动发送 --- 订阅模式
func TestRabbitMQSend(t *testing.T) {

	createCfg()

	content := make(map[string]interface{})
	content["name"] = "runThings"
	content["value"] = 1231

	req := &model.Eq2MqLog{
		Sn:       "1034639560",
		Product:  "qq群:925653309",
		Protocol: config.ProtocolUnknown,
		Status:   config.EqStatusOnline,
		Content:  content,
		Title:    "万物互联从此开始",
		Link:     false,
	}

	messageBytes, _ := json.Marshal(req) // 默认成功，不检测
	messageStr := string(messageBytes)
	fmt.Println("--- MQ发送内容：", messageStr)
	rabbitmq := service.NewRabbitMQPubSub(rabbitMqCfg.Exchange, rabbitMqCfg.Mqurl)
	rabbitmq.PublishSimple(messageStr)
}
