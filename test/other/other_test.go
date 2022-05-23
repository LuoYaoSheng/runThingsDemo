package other

import (
	"encoding/json"
	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	service "github.com/LuoYaoSheng/runThingsServer/core"
	"github.com/LuoYaoSheng/runThingsServer/extend"
	"log"
	"reflect"
	"testing"
)

func TestOther(t *testing.T) {

	sn := "tc_0001"
	code := "1100800013"

	objMap := map[string]float64{
		"temperature": 51.2,
		"humidity":    21.1,
	}

	dataType, _ := json.Marshal(objMap)

	var tempMap map[string]interface{}
	json.Unmarshal(dataType, &tempMap)

	status, title := dataCheck(sn, code, tempMap)
	log.Println(status, title)

}

func dataCheck(sn, code string, tempMap map[string]interface{}) (status_ int, title_ string) {

	// 初始化 redis 客户端
	service.GetRedisClient("127.0.0.1:6379", "123456", 0)

	objLists := extend.RuleFromRedis(sn, code)
	log.Println(objLists)

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
			case 7: //不在范围内
				vList := reflect.ValueOf(content.Value)
				if vList.Len() == 2 && !(objValue >= vList.Index(0).Float() && objValue <= vList.Index(1).Float()) {
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
