package main

import (
	"log"
	"time"

	"minlink/internal/config"
	"minlink/internal/handler"
	"minlink/internal/middleware"
	"minlink/internal/model"
	"minlink/internal/repository"
	"minlink/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/sqlite"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg := config.LoadConfig()

	db, err := gorm.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	// 自动迁移所有模型
	db.AutoMigrate(&model.ShortURL{}, &model.VisitLog{}, &model.User{})

	shortURLService := service.NewShortURLService(db)
	shortURLHandler := handler.NewShortURLHandler(shortURLService)

	// 用户管理服务和处理器
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	// 初始化默认管理员
	if err := userService.InitDefaultAdmin(); err != nil {
		log.Printf("Failed to init default admin: %v", err)
	}

	r := gin.Default()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// API 路由
	r.GET("/:shortCode", shortURLHandler.Redirect)

	api := r.Group("/api/v1")
	{
		// 公开接口
		api.POST("/short-url", shortURLHandler.CreateShortURL)
		api.GET("/short-url/:code/stats", shortURLHandler.GetStats)
		api.GET("/short-url/list", shortURLHandler.ListShortURLs)

		// 认证接口
		api.POST("/auth/login", userHandler.Login)

		// 需要登录的接口
		userGroup := api.Group("/user")
		userGroup.Use(middleware.JWTMiddleware(cfg.JWTSecret))
		{
			userGroup.GET("/profile", userHandler.GetProfile)
			userGroup.PUT("/password", userHandler.ChangePassword)
			userGroup.PUT("/profile", userHandler.UpdateProfile)
		}

		// 需要管理员权限的接口
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.JWTMiddleware(cfg.JWTSecret))
		adminGroup.Use(middleware.AdminMiddleware())
		{
			// 用户管理
			adminGroup.POST("/users", userHandler.CreateUser)
			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)

			// 短链管理（管理员可以管理所有短链）
			adminGroup.GET("/short-url/list", shortURLHandler.ListShortURLs)
			adminGroup.PUT("/short-url/:code/status", shortURLHandler.UpdateStatus)
			adminGroup.DELETE("/short-url/:code", shortURLHandler.DeleteShortURL)
		}
	}

	log.Printf("Server starting on port %s...", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
