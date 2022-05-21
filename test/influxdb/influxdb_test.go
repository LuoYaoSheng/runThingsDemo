package influxdb

import (
	"fmt"
	"testing"

	"github.com/LuoYaoSheng/runThingsConfig/config"
	"github.com/LuoYaoSheng/runThingsServer/core"
)

func TestInfluxdb(t *testing.T) {
	influxdbCfg := &config.InfluxdbConf{
		Addr:      "http://127.0.0.1:8086",
		Username:  "root",
		Password:  "root",
		Database:  "runThings",
		Precision: "",
		Prefix:    "test_",
	}

	sn := "89861119259042778836"
	imei := ""
	payload := make(map[string]interface{})

	payload["qq"] = "1034639560"
	payload["author"] = "寺西"

	service.GetClient(influxdbCfg.Addr, influxdbCfg.Username, influxdbCfg.Password, influxdbCfg.Database, influxdbCfg.Precision)

	_, err := service.WirteInflux(sn, imei, config.EqStatusNor, payload, influxdbCfg.Database, influxdbCfg.Prefix, influxdbCfg.Precision)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("写入成功")

	//sql := `where time > 1650384000000000000 and time < 1653062399000000000  LIMIT 30`
	sql := `where "status" = '0'` //索引列，key-value结构，value数据类型只支持string

	list, err := service.SelectInflux(sn, sql, influxdbCfg.Database, influxdbCfg.Prefix)
	if err != nil {
		return
	}
	fmt.Println("读取内容:", list)
}
