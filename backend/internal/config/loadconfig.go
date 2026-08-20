package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"` // 保留结构，但值会被环境变量覆盖
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load 加载 YAML 配置，但数据库配置强制从环境变量读取（忽略 YAML 中的值）
func Load(filename string) (Config, error) {
	// 1. 读取 YAML 文件（用于 Server、Redis、RabbitMQ 等非数据库配置）
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	// 2. 强制覆盖数据库配置：完全使用环境变量，无视 YAML
	cfg.Database = getDatabaseConfigFromEnv()

	// （可选）如果你也希望 Redis 和 RabbitMQ 强制使用环境变量，可以类似处理
	// cfg.Redis = getRedisConfigFromEnv()
	// cfg.RabbitMQ = getRabbitMQConfigFromEnv()

	return cfg, nil
}

// getDatabaseConfigFromEnv 从环境变量构建数据库配置，若无则使用默认值（本地开发）
func getDatabaseConfigFromEnv() DatabaseConfig {
	host := os.Getenv("MYSQLHOST")
	if host == "" {
		host = "localhost" // 本地默认
	}

	port := 3306
	if p := os.Getenv("MYSQLPORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	user := os.Getenv("MYSQLUSER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("MYSQLPASSWORD")
	if password == "" {
		password = "123456" // 本地开发默认
	}

	dbname := os.Getenv("MYSQLDATABASE")
	if dbname == "" {
		dbname = "feedsystem" // 本地默认
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}
