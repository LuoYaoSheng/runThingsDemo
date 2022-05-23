package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"log"
	"testing"

	"github.com/LuoYaoSheng/runThingsServer/core"
)

func receiveSimple(str string) {
	fmt.Println("---简单模式接收到数据: ", str)
}

func TestReceiveSimple(t *testing.T) {
	rabbitmq := service.NewRabbitMQSimple("runThings-heartbeat", "amqp://admin:admin@127.0.0.1:5672/")
	rabbitmq.ConsumeSimple(receiveSimple)
}

// 接收
func TestReceiveSimpleCmd(t *testing.T) {
	product := "p001"
	rabbitmq := service.NewRabbitMQSimple("runThings-cmd-"+product, "amqp://admin:admin@127.0.0.1:5672/")
	rabbitmq.ConsumeSimple(receiveSimple)
}

// 接收
func TestReceiveSimpleThreshold(t *testing.T) {

	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	// 初始化 rabbitmq
	rabbitmq := service.NewRabbitMQSimple("runThings-threshold", "amqp://admin:admin@127.0.0.1:5672/")
	rabbitmq.ConsumeSimple(receiveThreshold)
}

func receiveThreshold(str string) {
	var threshold model.Eq2MqThreshold
	err5 := json.Unmarshal([]byte(str), &threshold)
	if err5 != nil {
		log.Println(err5)
		return
	}
	switch threshold.Operate {
	case 0, 1:
		redisAddUpdateThreshold(threshold.Content)
	case 2:
		redisDelThreshold(threshold.Content)
	}
}

func redisAddUpdateThreshold(rule model.Rule) {
	key := rule.Code
	if len(key) <= 0 {
		key = rule.Sn
	}
	key = key + "_rule"
	value, _ := service.GetRdValue(key)
	if len(value) == 0 {
		saveValue := []model.Rule{rule}
		dataType, _ := json.Marshal(saveValue)
		err := service.SetRdValue(key, string(dataType))
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		var saveValue []model.Rule
		err := json.Unmarshal([]byte(value), &saveValue)
		if err != nil {
			log.Println(err)
			return
		}

		index := -1
		for idx, v := range saveValue {
			if v.Id == rule.Id {
				index = idx
				break
			}
		}
		if index < 0 {
			saveValue = append(saveValue, rule)
		} else {
			saveValue[index] = rule
		}
		dataType, _ := json.Marshal(saveValue)
		err = service.SetRdValue(key, string(dataType))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func redisDelThreshold(rule model.Rule) {
	objRule := rule
	key := objRule.Code
	if len(key) <= 0 {
		key = objRule.Sn
	}

	key = key + "_rule"
	value, _ := service.GetRdValue(key)

	var saveValue []model.Rule

	err := json.Unmarshal([]byte(value), &saveValue)
	if err != nil {
		log.Println(err)
		return
	}

	index := -1
	for idx, v := range saveValue {
		if v.Id == objRule.Id {
			index = idx
			break
		}
	}

	if index >= 0 {
		saveValue = append(saveValue[:index], saveValue[(index+1):]...)
	}
	dataType, _ := json.Marshal(saveValue)
	err = service.SetRdValue(key, string(dataType))
	if err != nil {
		log.Println(err)
		return
	}
}
