package rpc

// 模拟 RPC 进行接收

import (
	"encoding/json"
	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"
	"github.com/LuoYaoSheng/runThingsServer/extend"
	"log"
	"reflect"
	model2 "run-things-demo/eq/model"

	"strings"
)

var rabbitmqHeartbeat = service.NewRabbitMQSimple("runThings-heartbeat", "amqp://admin:admin@127.0.0.1:5672/")
var rabbitmqLog = service.NewRabbitMQSimple("runThings-logs", "amqp://admin:admin@127.0.0.1:5672/")
var redisClient = service.GetRedisClient("127.0.0.1:6379", "123456", 0) // 直接初始化

func Revive(topic, payload string) {

	log.SetFlags(log.Llongfile)

	//log.Println("rpc:", topic, payload)
	// 需要区分主题： 上报数据[update] / 心跳[heart] / 指令下发[cmd] / 指令应答[ack]
	topics := strings.Split(topic, "/")
	if len(topics) != 4 && topics[0] != "th-calc" {
		return // 过滤非标准
	}

	m := map[string]float64{
		"temperature": model2.TemperatureToplimit,
		"humidity":    model2.HumidityToplimit,
	}
	value, err := service.GetRdValue(topics[2] + "_m")
	if err == nil {
		// 获取值
		err2 := json.Unmarshal([]byte(value), &m)
		if err2 != nil {
			log.Println(err2)
		}
	}
	//log.Println("rpc-m: ", m)

	// 非下发指令，发送心跳
	if topics[3] != "cmd" {
		heart := model.Eq2MqHeartbeat{
			Sn:        topics[2],
			Product:   topics[1],
			Heartbeat: int64(model2.Heart),
		}
		msg, _ := json.Marshal(heart)
		rabbitmqHeartbeat.PublishSimple(string(msg))
	}

	// 非心跳包，上传数据到日志
	if topics[3] != "heart" && topics[3] != "cmd" && topics[3] != "ack" {
		var tempMap map[string]interface{}

		err = json.Unmarshal([]byte(payload), &tempMap)
		if err != nil {
			log.Println(err)
			return
		}

		status := config.EqStatusUnknown
		title := ""

		switch topics[3] {
		case "ack":
			status = config.EqStatusAck
		case "cmd":
			status = config.EqStatusCmd // 不用回传
		case "update":
			status, title = dataCheck(topics[1], topics[2], tempMap)
		}

		// 其他情况，发送日志
		mqLog := &model.Eq2MqLog{
			Sn:       topics[2],
			Product:  topics[1],
			Protocol: config.ProtocolMQTT,
			Status:   int64(status),
			Content:  tempMap,
			Title:    title,
			Link:     false,
		}

		msg, _ := json.Marshal(mqLog)
		rabbitmqLog.PublishSimple(string(msg))
	}

}

func dataCheck(sn, code string, tempMap map[string]interface{}) (status_ int, title_ string) {

	objLists := extend.RuleFromRedis(sn, code)
	//log.Println(objLists)

	// 返回内容预定义
	status := config.EqStatusNor
	title := ""

	temperature := tempMap["temperature"].(float64)
	humidity := tempMap["humidity"].(float64)

	for _, rule := range objLists {
		var contentList []model.RuleContent
		err := json.Unmarshal([]byte(rule.Content), &contentList)
		if err != nil {
			log.Println(err)
			return
		}
		checked := true
		for _, content := range contentList {
			//content.Property
			//log.Println(content)
			var objValue float64

			// 当前只有两种，暂时先这么写~~
			if content.Property == "temperature" {
				objValue = temperature
			}
			if content.Property == "humidity" {
				objValue = humidity
			}

			switch content.Condition {
			case 0: //大于
				if !(objValue > content.Value.(float64)) {
					checked = false
				}
			case 1: //大于等于
				if !(objValue >= content.Value.(float64)) {
					checked = false
				}
			case 2: //小于
				if !(objValue < content.Value.(float64)) {
					checked = false
				}
			case 3: //小于等于
				if !(objValue <= content.Value.(float64)) {
					checked = false
				}
			case 4: //等于
				if !(objValue == content.Value.(float64)) {
					checked = false
				}
			case 5: //不等于
				if !(objValue != content.Value.(float64)) {
					checked = false
				}
			case 6: //在范围内
				vList := reflect.ValueOf(content.Value)
				if vList.Len() == 2 && !(objValue < vList.Index(0).Float() && objValue > vList.Index(1).Float()) {
					checked = false
				}
				if vList.Len() != 2 {
					checked = false
				}
			case 7: //不在范围内
				vList := reflect.ValueOf(content.Value)
				if vList.Len() == 2 && !(objValue >= vList.Index(0).Float() && objValue <= vList.Index(1).Float()) {
					checked = false
				}
				if vList.Len() != 2 {
					checked = false
				}
			}
		}
		if checked {
			title = rule.Name
			status = config.EqStatusAlarm
		}
	}

	return status, title
}
