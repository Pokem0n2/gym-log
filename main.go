package main

import (
	"log"

	"github.com/Pokem0n2/gym-log/internal/config"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/Pokem0n2/gym-log/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	r := router.New(db)
	log.Printf("服务启动于 %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
