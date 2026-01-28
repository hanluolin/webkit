package config

import (
	"os"
	"strconv"
	"time"
	"webkit/kit/zlogger"
	"webkit/util"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

/**
 * config，配置包
   - 加载配置，两种方式：
     InitByEnv-通过环境变量加载配置
	 InitByFile-通过读取文件加载配置（支持yaml、json、toml等多种格式）
   - 取用配置：config.Conf.XX
*/

var Conf Config

// InitByEnv 通过环境变量加载配置
func InitByEnv() {
	Conf.Server = ServerConf{
		Port: GetEnv("SERVER_PORT", ":3000"),
	}
	Conf.DB = DBConf{
		Type: GetEnv("DB_TYPE", "pg"),
		Conn: GetEnv("DB_CONN", "host=127.0.0.1 port=5432 user=cella dbname=test password=111111"),
		LogLevel: GetEnv("DB_LOG_LEVEL", logger.Info),
	}
	Conf.Logger = zlogger.DefaultLog()
}

// InitByFile 通过文件加载配置
func InitByFile(fileName string) {
	viper.SetConfigFile(util.FindConfigFile(fileName))
	if err := viper.ReadInConfig(); err != nil {
		zap.S().Panic("config init fail", err)
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		zap.S().Panic("config init fail", err)
	}
	viper.Reset() // 释放viper内存
}

// GetEnv 获取环境变量
func GetEnv[T any](key string, defaultValue T) T {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value T
	switch any(defaultValue).(type) {
	case string:
		value = any(valueStr).(T)
	case int:
		v, err := strconv.Atoi(valueStr)
		if err != nil {
			return defaultValue
		}
		value = any(v).(T)
	case bool:
		v, err := strconv.ParseBool(valueStr)
		if err != nil {
			return defaultValue
		}
		value = any(v).(T)
	case float64:
		v, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return defaultValue
		}
		value = any(v).(T)
	default:
		return defaultValue
	}
	return value
}

type Config struct {
	Server ServerConf
	Logger *zlogger.Option
	DB     DBConf
	Redis  *redis.Options
}

type ServerConf struct {
	Port string
}

type DBConf struct {
	Type          string          `json:"type"`
	Conn          string          `json:"conn"`
	MaxIdleConn   int             `json:"max_idle_conn"`   // 最大空闲连接
	MaxOpenConn   int             `json:"max_open_conn"`   // 最大连接数
	MaxLifeTime   int             `json:"max_life_time"`   // 最大活跃时间，单位：h
	MaxIdleTime   int             `json:"max_idle_time"`   // 最大空闲保活时间，单位：h
	SlowQueryTime time.Duration   `json:"slow_query_time"` // 慢 SQL 阈值
	LogLevel      logger.LogLevel `json:"log_level"`       // 日志等级
	LogColorful   bool            `json:"log_colorful"`    // 启用彩色日志
}
