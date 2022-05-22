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

type Rule struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Code    string `json:"code"`
	Sn      string `json:"sn"`
	Content string `json:"content"`
}

var rules = []Rule{
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
			saveValue := []Rule{rules[i]}
			dataType, _ := json.Marshal(saveValue)
			err := service.SetRdValue(key, string(dataType))
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			var saveValue []Rule
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

	var saveValue []Rule

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

	var key string
	// 获取 sn 对应规则
	key = sn + "_rule"
	snValue, _ := service.GetRdValue(key)
	var snRules []Rule
	if len(snValue) > 0 {
		err := json.Unmarshal([]byte(snValue), &snRules)
		if err != nil {
			log.Println(err)
			return
		}
	}
	//log.Println(snRules)

	// 获取 code 对应规则
	key = code + "_rule"
	codeValue, _ := service.GetRdValue(key)
	var codeRules []Rule
	if len(codeValue) > 0 {
		err := json.Unmarshal([]byte(codeValue), &codeRules)
		if err != nil {
			log.Println(err)
			return
		}
	}
	//log.Println(codeRules)

	rules_ := append(snRules, codeRules...) // 一定要 snRules在前，重复时好保留
	log.Println(rules_)
	objRules := RemoveRepByLoop(rules_)
	log.Println(objRules)
}

type RuleContent struct {
	Property  string      `json:"property"`
	Condition int         `json:"condition"`
	Value     interface{} `json:"value"`
}

// 通过两重循环过滤重复元素
func RemoveRepByLoop(slc []Rule) []Rule {
	result := []Rule{} // 存放结果
	for i := range slc {
		flag := true
		for j := range result {

			log.Println(slc[i].Content)
			slcMap := []RuleContent{}
			err := json.Unmarshal([]byte(slc[i].Content), &slcMap)
			if err != nil {
				log.Println(err)
				break
			}

			resMap := []RuleContent{}
			err = json.Unmarshal([]byte(result[j].Content), &resMap)
			if err != nil {
				log.Println(err)
				break
			}

			if len(slcMap) == len(resMap) {
				flag2 := true
				for k := 0; k < len(slcMap); k++ {
					if !(slcMap[k].Property == resMap[k].Property && slcMap[k].Condition == resMap[k].Condition) {
						flag2 = false
						break
					}
				}
				if flag2 == true {
					flag = false // 存在重复元素，标识为false
					break
				}
			}

		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}
