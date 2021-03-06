package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"
)

// 测试发送实时日志
func TestSendSimple(t *testing.T) {
	// 发送多个，便于测试工厂模式
	rabbitmq := service.NewRabbitMQSimple("runThings-logs", "amqp://admin:admin@127.0.0.1:5672/")

	content := map[string]interface{}{
		"name":    "runThings",
		"version": 1.0,
		"content": "一个小型IoT管理中间件",
	}

	mqLog := model.Eq2MqLog{
		Sn:       "1034639560",
		Product:  "p001",
		Protocol: 0,
		Status:   0,
		Content:  content,
		Title:    "TestSendSimple",
		Link:     false,
	}

	for i := 0; i < 10; i++ {
		mqLog.Status = int64(i)
		dataType, _ := json.Marshal(mqLog)
		// 发送 log
		//msg := `{"sn":"1034639560","product":"p001","protocol":0,"status":` + strconv.Itoa(i) + `,"content":{"name":"runThings","value":1231},"title":"万物互联从此开始","link":false}`
		msg := string(dataType)
		rabbitmq.PublishSimple(msg)
	}
}

// 测试发送心跳包
func TestSendSimple2(t *testing.T) {
	// 发送 设备上下线
	rabbitmq := service.NewRabbitMQSimple("runThings-heartbeat", "amqp://admin:admin@127.0.0.1:5672/")

	heartbeat := 10

	m := model.Eq2MqHeartbeat{
		Sn:        "1034639560",
		Product:   "p001",
		Heartbeat: int64(heartbeat),
	}
	dataType, _ := json.Marshal(m)

	//msg := `{"sn":"1034639560","product":"p001", "heartbeat":` + strconv.Itoa(heartbeat) + `}`
	msg := string(dataType)

	// 6 * 10 等于 1分钟
	for i := 0; i < 6; i++ {
		rabbitmq.PublishSimple(msg)
		// 模拟心跳间隔
		fmt.Println("发送心跳", msg)
		time.Sleep(time.Duration(heartbeat) * time.Second)
	}
}

func TestThreshold(t *testing.T) {
	rabbitmq := service.NewRabbitMQSimple("runThings-threshold", "amqp://admin:admin@127.0.0.1:5672/")

	rule := model.Rule{
		Id:      1,
		Name:    "温度过高",
		Level:   1,
		Code:    "p_001",
		Sn:      "",
		Content: `[{"property":"temperature","condition":0,"value":70}]`,
	}

	mqLog := model.Eq2MqThreshold{
		Operate: 0,
		Content: rule,
	}

	dataType, _ := json.Marshal(mqLog)
	msg := string(dataType)

	log.Print(msg)
	rabbitmq.PublishSimple(msg)
}
