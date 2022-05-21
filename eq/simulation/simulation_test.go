package simulation

// 模拟一台设备
import (
	"encoding/json"
	"log"
	"run-things-demo/eq/model"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	model2 "github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"math/rand"
	"testing"
	"time"
)

var on = true

var simulationMessagePubHandler mqtt.MessageHandler = func(mqttClient mqtt.Client, msg mqtt.Message) {
	//fmt.Printf("simulationMessagePubHandler: %s from topic: %s\n", msg.Payload(), msg.Topic())

	cmd := &model2.Eq2MqCmd{}
	err := json.Unmarshal(msg.Payload(), &cmd)
	if err != nil {
		log.Panicln(err)
		return
	}

	on = cmd.Content["on"].(bool)
	if on {
		log.Println("接收到命令：打开开关")
	} else {
		log.Println("接收到命令：关闭开关")
	}

	// 发送一个设备应答
	topic := `th-calc/` + model.ProductKey + `/` + cmd.Sn + `/ack`
	content := ""
	err = service.MqttSend(topic, content, mqttCfg.Qos)
	if err != nil {
		log.Panicln(err)
	}
}
var mqttCfg *config.MqttConf

func TestSimulation(t *testing.T) {

	pkey := model.ProductKey
	rand.Seed(time.Now().Unix())
	//sn := "tc_" + fmt.Sprintf("%03d", rand.Intn(10)+1) // 设置动态，可以开启多个设备
	sn := "tc_" + "0001"

	mqttCfg = &config.MqttConf{
		Broker:   "ws://127.0.0.1:8083/mqtt",
		Username: "",
		Password: "",
		Topic:    `th-calc/` + pkey + `/` + sn + `/cmd`,
		Qos:      0,
	}

	// 开启订阅模式 -- 会初始化 mqtt客户端
	err := service.MqttSubscribe(mqttCfg.Broker, mqttCfg.Username, mqttCfg.Password, mqttCfg.Topic, mqttCfg.Qos, simulationMessagePubHandler)
	if err != nil {
		log.Panicln(err)
		return
	}

	// 心跳包
	go func() {
		for true {
			topic := `th-calc/` + pkey + `/` + sn + `/heart`
			content := ""
			err = service.MqttSend(topic, content, mqttCfg.Qos)
			if err != nil {
				log.Panicln(err)
			}

			time.Sleep(time.Duration(model.Heart) * time.Second)
		}
	}()

	// 业务数据包
	go func() {
		for true {
			//rand.Seed(time.Now().Unix())

			m := model.ThCalcModel{
				Temperature: rand.Float64() * 100,
				Humidity:    rand.Float64() * 100,
				On:          on,
				CurTime:     time.Now(),
			}

			content, _ := json.Marshal(m)

			topic := `th-calc/` + pkey + `/` + sn + `/update`
			err = service.MqttSend(topic, content, mqttCfg.Qos)
			if err != nil {
				log.Panicln(err)
			}
			// 随机等待时间
			//time.Sleep(time.Duration(rand.Intn(10)+1) * time.Second) // 随机会出现，先过滤
			time.Sleep(3 * time.Second) // 每3秒一次
		}
	}()

	select {} // 强制等待
}
