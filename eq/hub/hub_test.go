package hub

// 模拟一个业务中心
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	model2 "run-things-demo/eq/model"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/LuoYaoSheng/runThingsServer/core"

	"strconv"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	rabbitmqCmd       *service.RabbitMQ
	rabbitmqThreshold *service.RabbitMQ
	db                *sqlx.DB
)

var MessagePubHandler mqtt.MessageHandler = func(mqttClient mqtt.Client, msg mqtt.Message) {
	fmt.Printf("+++++++++++Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	topics := strings.Split(msg.Topic(), "/")
	if len(topics) != 5 {
		log.Panicln("过滤非标准")
		return
	}

	sn := topics[3]
	productKey := topics[2]
	status, _ := strconv.Atoi(topics[4])

	mgLog := &model.Eq2MqLog{}
	err := json.Unmarshal(msg.Payload(), &mgLog)
	if err != nil {
		log.Panicln(err)
		return
	}
	content, _ := json.Marshal(mgLog.Content)

	// 存储表 -- 需要考虑异常事件重复问题: 可能又要用到Redis，压力挺大
	InsertMySQL(sn, productKey, mgLog.Title, string(content), status)

	// 业务端发送告警到前端

	msf := map[string]interface{}{}
	err = json.Unmarshal(content, &msf)
	if err != nil {
		log.Panicln(err)
		return
	}
}

func InsertMySQL(sn, productKey, title, content string, status int) {
	sql := "insert into eq_log (sn,product_key, status,title, content,create_time)values (?,?,?,?,?,?)"
	value := [6]interface{}{sn, productKey, status, title, content, time.Now()}

	//执行SQL语句
	_, err := db.Exec(sql, value[0], value[1], value[2], value[3], value[4], value[5])
	if err != nil {
		log.Println("exec failed,", err)
	}
}

func GetMySqlAndMQ() {

	var rules []model.Rule

	sql := `select id,name,level,code,sn,content from eq_alarm_rule`
	err := db.Select(&rules, sql)
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(rules)

	// 发送rabbitmq 存储到 Redis 中
	for _, rule := range rules {
		threshold := model.Eq2MqThreshold{
			Operate: 0,
			Content: rule,
		}
		thresholdMQ(threshold)
	}
}

func cmdMQ(cmd model.Eq2MqCmd) {
	data, _ := json.Marshal(cmd)
	log.Println("---- 下发指令 ----", string(data))
	rabbitmqCmd.PublishSimple(string(data))
}

func thresholdMQ(threshold model.Eq2MqThreshold) {
	data, _ := json.Marshal(threshold)
	//log.Println("---- 设置告警规则 ----", string(data))
	rabbitmqThreshold.PublishSimple(string(data))
}

type server int

func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.URL.Path)

	onValue := r.URL.Query().Get("on")
	snValue := r.URL.Query().Get("sn")

	if len(snValue) <= 0 {
		log.Println("sn 未设置")
		return
	}

	on := true
	if onValue == `0` || onValue == `false` || len(onValue) == 0 {
		on = false
	}

	cmdContent := map[string]interface{}{}
	cmdContent["on"] = on

	cmd := model.Eq2MqCmd{
		Sn:      snValue,
		Content: cmdContent,
	}

	cmdMQ(cmd)

	w.Write([]byte("指令已下发"))
}

func TestHub(t *testing.T) {

	log.SetFlags(log.Llongfile)

	// 订阅 MQTT ，获取设备异常情况
	topic := "event/runTings/" + model2.ProductKey + "/#"
	mqttCfg := &config.MqttConf{
		Broker: "ws://127.0.0.1:8083/mqtt",
		Topic:  topic,
	}

	// 开启订阅模式
	err := service.MqttSubscribe(mqttCfg.Broker, mqttCfg.Username, mqttCfg.Password, mqttCfg.Topic, mqttCfg.Qos, MessagePubHandler)
	if err != nil {
		log.Panicln(err)
		return
	}

	// 获取 Redis
	redisCfg := &config.RedisConf{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	}
	service.GetRedisClient(redisCfg.Addr, redisCfg.Password, redisCfg.DB)

	// 获取 rabbitmq
	rabbitmqCmd = service.NewRabbitMQSimple("runThings-cmd-"+model2.ProductKey, "amqp://admin:admin@127.0.0.1:5672/")
	rabbitmqThreshold = service.NewRabbitMQSimple("runThings-threshold", "amqp://admin:admin@127.0.0.1:5672/")

	// 获取 mysql
	database, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/eq")
	if err != nil {
		log.Panicln("open mysql failed,", err)
	}
	db = database

	// 获取告警规则
	GetMySqlAndMQ()

	// 开启http
	var s server
	err = http.ListenAndServe(":9999", &s)
	if err != nil {
		log.Panicln("open http failed,", err)
		return
	}

}
