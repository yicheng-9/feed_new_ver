package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
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

// Load 从 YAML 文件加载配置，并用环境变量覆盖（环境变量优先级更高）
func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	// ---------- 数据库配置 ----------
	// 1. 优先尝试解析 MYSQL_URL
	if mysqlURL := os.Getenv("MYSQL_URL"); mysqlURL != "" {
		if parsed, err := url.Parse(mysqlURL); err == nil {
			if parsed.User != nil {
				cfg.Database.User = parsed.User.Username()
				if pass, ok := parsed.User.Password(); ok {
					cfg.Database.Password = pass
				}
			}
			cfg.Database.Host = parsed.Hostname()
			if port := parsed.Port(); port != "" {
				if p, err := strconv.Atoi(port); err == nil {
					cfg.Database.Port = p
				}
			}
			if len(parsed.Path) > 1 {
				cfg.Database.DBName = strings.TrimPrefix(parsed.Path, "/")
			}
		}
	} else {
		// 2. 否则尝试读取 Railway 提供的分开变量（无下划线版本）
		if v := os.Getenv("MYSQLHOST"); v != "" {
			cfg.Database.Host = v
		} else if v := os.Getenv("DB_HOST"); v != "" {
			cfg.Database.Host = v
		}
		if v := os.Getenv("MYSQLPORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.Database.Port = p
			}
		} else if v := os.Getenv("DB_PORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.Database.Port = p
			}
		}
		// MYSQLUSER 或 DB_USER（注意 Railway 可能是 MYSQLUSER 或 MYSQL_USER，这里同时支持）
		if v := os.Getenv("MYSQLUSER"); v != "" {
			cfg.Database.User = v
		} else if v := os.Getenv("MYSQL_USER"); v != "" {
			cfg.Database.User = v
		} else if v := os.Getenv("DB_USER"); v != "" {
			cfg.Database.User = v
		}
		if v := os.Getenv("MYSQLPASSWORD"); v != "" {
			cfg.Database.Password = v
		} else if v := os.Getenv("MYSQL_PASSWORD"); v != "" {
			cfg.Database.Password = v
		} else if v := os.Getenv("DB_PASSWORD"); v != "" {
			cfg.Database.Password = v
		}
		if v := os.Getenv("MYSQL_DATABASE"); v != "" {
			cfg.Database.DBName = v
		} else if v := os.Getenv("DB_NAME"); v != "" {
			cfg.Database.DBName = v
		}
	}

	// ---------- Redis 配置 ----------
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if parsed, err := url.Parse(redisURL); err == nil {
			if parsed.User != nil {
				if pass, ok := parsed.User.Password(); ok {
					cfg.Redis.Password = pass
				}
			}
			cfg.Redis.Host = parsed.Hostname()
			if port := parsed.Port(); port != "" {
				if p, err := strconv.Atoi(port); err == nil {
					cfg.Redis.Port = p
				}
			}
			if len(parsed.Path) > 1 {
				if db, err := strconv.Atoi(strings.TrimPrefix(parsed.Path, "/")); err == nil {
					cfg.Redis.DB = db
				}
			}
		}
	} else {
		if v := os.Getenv("REDISHOST"); v != "" {
			cfg.Redis.Host = v
		} else if v := os.Getenv("REDIS_HOST"); v != "" {
			cfg.Redis.Host = v
		}
		if v := os.Getenv("REDISPORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.Redis.Port = p
			}
		} else if v := os.Getenv("REDIS_PORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.Redis.Port = p
			}
		}
		if v := os.Getenv("REDISPASSWORD"); v != "" {
			cfg.Redis.Password = v
		} else if v := os.Getenv("REDIS_PASSWORD"); v != "" {
			cfg.Redis.Password = v
		}
		if v := os.Getenv("REDISDB"); v != "" {
			if db, err := strconv.Atoi(v); err == nil {
				cfg.Redis.DB = db
			}
		} else if v := os.Getenv("REDIS_DB"); v != "" {
			if db, err := strconv.Atoi(v); err == nil {
				cfg.Redis.DB = db
			}
		}
	}

	// ---------- RabbitMQ 配置 ----------
	if rabbitURL := os.Getenv("RABBITMQ_URL"); rabbitURL != "" {
		if parsed, err := url.Parse(rabbitURL); err == nil {
			if parsed.User != nil {
				cfg.RabbitMQ.Username = parsed.User.Username()
				if pass, ok := parsed.User.Password(); ok {
					cfg.RabbitMQ.Password = pass
				}
			}
			cfg.RabbitMQ.Host = parsed.Hostname()
			if port := parsed.Port(); port != "" {
				if p, err := strconv.Atoi(port); err == nil {
					cfg.RabbitMQ.Port = p
				}
			}
		}
	} else {
		if v := os.Getenv("RABBITMQ_HOST"); v != "" {
			cfg.RabbitMQ.Host = v
		} else if v := os.Getenv("RABBITMQHOST"); v != "" {
			cfg.RabbitMQ.Host = v
		}
		if v := os.Getenv("RABBITMQ_PORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.RabbitMQ.Port = p
			}
		} else if v := os.Getenv("RABBITMQPORT"); v != "" {
			if p, err := strconv.Atoi(v); err == nil {
				cfg.RabbitMQ.Port = p
			}
		}
		if v := os.Getenv("RABBITMQ_USERNAME"); v != "" {
			cfg.RabbitMQ.Username = v
		} else if v := os.Getenv("RABBITMQUSERNAME"); v != "" {
			cfg.RabbitMQ.Username = v
		}
		if v := os.Getenv("RABBITMQ_PASSWORD"); v != "" {
			cfg.RabbitMQ.Password = v
		} else if v := os.Getenv("RABBITMQPASSWORD"); v != "" {
			cfg.RabbitMQ.Password = v
		}
	}

	// ---------- Server 端口 ----------
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = p
		}
	}

	return cfg, nil
}
