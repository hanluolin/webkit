package model

import (
	"errors"
	"log"
	"strings"
	"time"
	"webkit/config"

	"github.com/LorinHan/dm-driver/dmgorm2"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func Init(dbConf config.DBConf) error {
	var (
		err  error
		conf = &gorm.Config{
			TranslateError:         true,
			SkipDefaultTransaction: true,
		}
		logConf = logger.Config{
			SlowThreshold: dbConf.SlowQueryTime,
			LogLevel:      dbConf.LogLevel,
			Colorful:      dbConf.LogColorful,
		}
	)

	if dbConf.LogLevel == 0 {
		logConf.LogLevel = logger.Info // 默认Warn级别（慢查询）
	}
	conf.Logger = logger.New(log.New(&GormLogger{}, "", 0), logConf)

	switch dbConf.Type {
	case "pg":
		db, err = gorm.Open(postgres.Open(dbConf.Conn), conf)
	case "dm":
		db, err = gorm.Open(dmgorm2.Open(dbConf.Conn), conf)
	default:
		zap.S().Fatal("unknown db type：", dbConf.Type)
	}
	if err != nil {
		return err
	}

	dbPool, err := db.DB()
	if err != nil {
		return err
	}

	// 设置空闲连接池中连接的最大数量
	dbPool.SetMaxIdleConns(dbConf.MaxIdleConn)

	// 设置打开数据库连接的最大数量
	dbPool.SetMaxOpenConns(dbConf.MaxOpenConn)

	// 设置连接可复用的最大时间
	if dbConf.MaxLifeTime != 0 {
		dbPool.SetConnMaxLifetime(time.Duration(dbConf.MaxLifeTime) * time.Hour)
	}
	// 设置连接处于空闲状态时可复用的最大时间
	if dbConf.MaxIdleTime != 0 {
		dbPool.SetConnMaxIdleTime(time.Duration(dbConf.MaxIdleTime) * time.Hour)
	}

	return errCallbackRegister()
}

type GormLogger struct {
}

func (*GormLogger) Write(p []byte) (n int, err error) {
	zap.S().WithOptions(zap.WithCaller(false)).Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

// errCallbackRegister 注册err处理回调
func errCallbackRegister() error {
	if err := db.Callback().Query().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	if err := db.Callback().Create().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	if err := db.Callback().Update().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	if err := db.Callback().Row().After("*").Register("err_callback", ErrorCallback); err != nil {
		return err
	}
	return nil
}

type DBErr struct {
	err error
}

func (d DBErr) Error() string {
	return d.err.Error()
}

func ErrorCallback(edb *gorm.DB) {
	if edb.Error == nil {
		return
	}
	if errors.Is(edb.Error, gorm.ErrRecordNotFound) {
		return
	}
	edb.Error = DBErr{err: edb.Error}
}
