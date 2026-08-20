package main

import (
	"context"
	"feedsystem_video_go/internal/config"
	"feedsystem_video_go/internal/db"
	apphttp "feedsystem_video_go/internal/http"
	rabbitmq "feedsystem_video_go/internal/middleware/rabbitmq"
	rediscache "feedsystem_video_go/internal/middleware/redis"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)
func getDatabaseConfigFromEnv() config.DatabaseConfig {
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
	// 加载配置（用于 Server、Redis、RabbitMQ 等其他配置）
	log.Printf("Loading config from configs/config.yaml")
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 直接从环境变量构建数据库 DSN，忽略 cfg.Database
	dbDSN := getDBDSNFromEnv()
	log.Printf("Database DSN: %s", dbDSN) // 打印但注意隐藏密码，生产上可去掉

	// 使用 dbDSN 连接数据库
	// 你需要修改 db.NewDB 函数，使其接受 DSN 字符串而非 config.DatabaseConfig
	// 或者我们临时在 db 包中添加一个 NewDBWithDSN 函数。
	// 为了快速解决，我们修改 db.NewDB 函数，使其接受 DSN（见下文修改 db.go）
	
	// 由于我们不想改动太多，我们直接修改 db.NewDB 函数让它接受 DSN，或在这里构造 config.DatabaseConfig 结构体覆盖 cfg.Database
	// 更简单：将 cfg.Database 赋值为从环境变量读取的值
	cfg.Database = getDatabaseConfigFromEnv()

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
