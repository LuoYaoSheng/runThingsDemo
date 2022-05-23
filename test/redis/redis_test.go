package redis

import (
	"encoding/json"
	"fmt"
	service "github.com/LuoYaoSheng/runThingsServer/core"
	"github.com/LuoYaoSheng/runThingsServer/extend"
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

	content := model.Rule{
		Id:      1,
		Name:    "测试",
		Level:   0,
		Code:    "",
		Sn:      "QQ_1034639560",
		Content: "[]",
	}

	threshold := model.Eq2MqThreshold{
		Operate: 0,
		Content: content,
	}

	dataType, _ := json.Marshal(threshold.Content)

	err := service.SetRdValue(content.Sn, string(dataType))
	if err != nil {
		log.Panic(err)
		return
	}
}

var rules = []model.Rule{
	{
		Id:      1,
		Name:    "温度过高",
		Level:   1,
		Code:    "p_001",
		Sn:      "",
		Content: `[{"property":"temperature","condition":0,"value":70}]`,
	},
	{
		Id:      2,
		Name:    "温度|湿度异常",
		Level:   0,
		Code:    "p_001",
		Sn:      "",
		Content: `[{"property":"temperature","condition":0,"value":50},{"property":"humidity","condition":0,"value":60}]`,
	},
	{
		Id:      3,
		Name:    "温度过高",
		Level:   1,
		Code:    "",
		Sn:      "sn_001",
		Content: `[{"property":"temperature","condition":0,"value":50}]`,
	},
}

// 新增或更新
func TestRedisAdd(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	// 新增或更新
	for i := 0; i < 3; i++ {
		key := rules[i].Code
		if len(key) <= 0 {
			key = rules[i].Sn
		}
		key = key + "_rule"
		value, _ := service.GetRdValue(key)
		if len(value) == 0 {
			saveValue := []model.Rule{rules[i]}
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
				if v.Id == rules[i].Id {
					index = idx
					break
				}
			}
			if index < 0 {
				saveValue = append(saveValue, rules[i])
			} else {
				saveValue[index] = rules[i]
			}
			dataType, _ := json.Marshal(saveValue)
			err = service.SetRdValue(key, string(dataType))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

}

// 删除
func TestRedisDel(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)
	objRule := rules[1]
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

// 获取执行项目
func TestRedisRun(t *testing.T) {
	log.SetFlags(log.Llongfile)
	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	sn := rules[2].Sn
	code := rules[0].Code

	objLists := extend.RuleFromRedis(sn, code)
	log.Println(objLists)
}
