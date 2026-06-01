package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"shipping-excel/backend/internal/db"
	"shipping-excel/backend/internal/handler"
	"shipping-excel/backend/internal/logx"
	"shipping-excel/backend/internal/service"
)

func main() {
	dataDir := envOr("DATA_DIR", "./data")
	port := envOr("PORT", "8099")

	database, err := db.Init(dataDir)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	jobSvc := service.NewJobService(database, dataDir)
	h := handler.New(jobSvc)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: false,
	}))

	h.Register(r)

	logx.Infof("server starting port=%s data_dir=%s", port, dataDir)
	log.Printf("服务启动于 :%s, 数据目录: %s", port, dataDir)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
