package redis

import (
	"encoding/json"
	"fmt"
	"github.com/LuoYaoSheng/runThingsServer/core"
	"log"

	"testing"
	"time"

	"github.com/LuoYaoSheng/runThingsConfig/model"
)

func testReviceMsg(str string) {
	fmt.Println(str, "解散通知~~~")
}

func TestRedis(t *testing.T) {

	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	// 设置过期
	expiration := time.Duration(1) * time.Second
	err := service.SetRdValueTimeout("QQ群", "925653309", expiration)
	if err != nil {
		log.Panicln(err)
		return
	}

	// 过期订阅
	err1 := service.SubscribeKeyExpired(testReviceMsg)
	if err1 != nil {
		log.Panicln(err1)
		return
	}

	time.Sleep(expiration + 1*time.Second)
}

func TestRedisThreshold(t *testing.T) {
	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	content := map[string]interface{}{
		"txt":     "string",
		"int":     1,
		"float64": 1.2,
	}

	threshold := model.Eq2MqThreshold{
		Sn:      "1111",
		Content: content,
	}

	dataType, _ := json.Marshal(threshold.Content)

	err := service.SetRdValue(threshold.Sn, string(dataType))
	if err != nil {
		log.Panic(err)
		return
	}
}