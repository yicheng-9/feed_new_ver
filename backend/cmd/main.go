package main

import (
	"context"
	"feedsystem_video_go/internal/config"
	"feedsystem_video_go/internal/db"
	apphttp "feedsystem_video_go/internal/http"
	rabbitmq "feedsystem_video_go/internal/middleware/rabbitmq"
	rediscache "feedsystem_video_go/internal/middleware/redis"
	"log"
	"os"
	"strconv"
	"time"
	"net/url"
)
func getDBConfigFromEnv() config.DatabaseConfig {
	// 1. 优先使用 MYSQL_URL（Railway 一定会提供）
	if mysqlURL := os.Getenv("MYSQL_URL"); mysqlURL != "" {
		// 格式: mysql://user:pass@host:port/dbname
		parsed, err := url.Parse(mysqlURL)
		if err == nil {
			cfg := config.DatabaseConfig{}
			if parsed.User != nil {
				cfg.User = parsed.User.Username()
				if pass, ok := parsed.User.Password(); ok {
					cfg.Password = pass
				}
			}
			cfg.Host = parsed.Hostname()
			if port := parsed.Port(); port != "" {
				if p, err := strconv.Atoi(port); err == nil {
					cfg.Port = p
				}
			}
			if len(parsed.Path) > 1 {
				cfg.DBName = parsed.Path[1:]
			}
			// 如果解析成功，直接返回
			if cfg.Host != "" && cfg.User != "" {
				return cfg
			}
		}
		// 如果解析失败，继续尝试分开变量
	}

	// 2. 回退到分开变量（用于本地开发或其他平台）
	host := os.Getenv("MYSQLHOST")
	if host == "" {
		host = "localhost"
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
		password = "123456"
	}
	dbname := os.Getenv("MYSQLDATABASE")
	if dbname == "" {
		dbname = "feedsystem"
	}
	return config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}
func main() {
	log.Println("===== NEW VERSION WITH ENV OVERRIDE =====")

	// 加载配置
	log.Printf("Loading config from configs/config.yaml")
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 强制使用环境变量覆盖数据库配置
	cfg.Database = getDBConfigFromEnv()
	log.Printf("Using database config: host=%s port=%d user=%s dbname=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.DBName)

	// 连接数据库
	sqlDB, err := db.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(sqlDB); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	defer db.CloseDB(sqlDB)

	// 连接 Redis (可选，用于缓存)
	cache, err := rediscache.NewFromEnv(&cfg.Redis)
	if err != nil {
		log.Printf("Redis config error (cache disabled): %v", err)
		cache = nil
	} else {
		pingCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		if err := cache.Ping(pingCtx); err != nil {
			log.Printf("Redis not available (cache disabled): %v", err)
			_ = cache.Close()
			cache = nil
		} else {
			defer cache.Close()
			log.Printf("Redis connected (cache enabled)")
		}
	}

	// 连接 RabbitMQ (可选，用于消息队列)
	rmq, err := rabbitmq.NewRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		log.Printf("RabbitMQ config error (disabled): %v", err)
		rmq = nil
	} else {
		defer rmq.Close()
		log.Printf("RabbitMQ connected")
	}

	// 设置路由
	r := apphttp.SetRouter(sqlDB, cache, rmq)
	log.Printf("Server is running on port %d", cfg.Server.Port)
	if err := r.Run(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
