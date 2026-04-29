package main

import (
	"flag"
	"log"
	"os"

	"github.com/Pokem0n2/gym-log/internal/config"
	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/Pokem0n2/gym-log/internal/router"
)

func main() {
	// 管理员子命令
	var (
		adminCreate = flag.Bool("admin-create", false, "创建管理员账号")
		userID      = flag.String("user-id", "", "用户ID，格式 usr_xxxx")
		username    = flag.String("username", "", "用户名")
		password    = flag.String("password", "", "密码")
	)
	flag.Parse()

	cfg := config.Load()

	// 设置 JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "gym-log-default-secret-change-me"
	}
	middleware.SetJWTSecret(jwtSecret)

	store, err := repository.NewUserStore(cfg.UserDataDir)
	if err != nil {
		log.Fatalf("用户数据库初始化失败: %v", err)
	}
	defer store.Close()

	// 管理员创建账号模式
	if *adminCreate {
		if *userID == "" || *username == "" || *password == "" {
			log.Fatal("缺少参数: -user-id, -username, -password 均为必填")
		}
		if err := store.CreateUser(*userID, *username, *password); err != nil {
			log.Fatalf("创建账号失败: %v", err)
		}
		log.Printf("账号创建成功: %s (%s)", *userID, *username)
		return
	}

	r := router.New(store)
	log.Printf("服务启动于 %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
