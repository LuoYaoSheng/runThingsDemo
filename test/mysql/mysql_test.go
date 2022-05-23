package mysql

import (
	"fmt"
	"github.com/LuoYaoSheng/runThingsConfig/model"
	"github.com/jmoiron/sqlx"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {

	database, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/eq")
	if err != nil {
		log.Panicln("open mysql failed,", err)

	}
	sql := `select id,name,level,code,sn,content from eq_alarm_rule`
	rows, err := database.Queryx(sql)

	if err != nil {
		log.Panicln(err)
	}

	rule := model.Rule{}

	for rows.Next() {
		err = rows.StructScan(&rule)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", rule)
	}
}

func TestMySQLSelect(t *testing.T) {

	db, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/eq")
	if err != nil {
		log.Panicln("open mysql failed,", err)

	}

	var rules []model.Rule

	sql := `select id,name,level,code,sn,content from eq_alarm_rule`
	err = db.Select(&rules, sql)
	if err != nil {
		log.Panicln(err)
		return
	}

	log.Println(rules)

	log.Println(rules[0].Name)

}
