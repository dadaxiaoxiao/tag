package ioc

import (
	"github.com/dadaxiaoxiao/tag/internal/repository/dao"
	promsdk "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// InitDB 根据key 初始化数据库
func InitDB() *gorm.DB {
	// username:password@protocol(address)/dbname
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var config Config
	err := viper.UnmarshalKey("db", &config)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
	if err != nil {
		// panic 相当于goroutine 结束
		panic(err)
	}

	cb := newCallbacks()
	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	// 生成数据表结构
	return db
}

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func (c *Callbacks) Name() string {
	return "prometheus-query"
}

func (c *Callbacks) Initialize(db *gorm.DB) error {
	c.registerAll(db)
	return nil
}

func newCallbacks() *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: "qinye_yiyi",
		Subsystem: "demo_tag",
		Name:      "gorm_query_time",
		Help:      "统计 GORM 执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	cb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return cb
}

func (c *Callbacks) registerAll(db *gorm.DB) {
	db.Callback().Create().Before("*").
		Register("promethues_create_before", c.before())
	db.Callback().Create().After("*").
		Register("promethues_create_after", c.after("create"))

	db.Callback().Update().Before("*").
		Register("promethues_update_before", c.before())
	db.Callback().Update().After("*").
		Register("promethues_update_after", c.after("update"))

	db.Callback().Delete().Before("*").
		Register("promethues_delete_before", c.before())
	db.Callback().Delete().After("*").
		Register("promethues_delete_after", c.after("delete"))

	db.Callback().Raw().Before("*").
		Register("promethues_raw_before", c.before())
	db.Callback().Raw().After("*").
		Register("promethues_raw_after", c.after("raw"))

	db.Callback().Row().Before("*").
		Register("promethues_row_before", c.before())
	db.Callback().Row().After("*").
		Register("promethues_row_after", c.after("row"))
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		// 类型断言
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		// 上报prometheus
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}
