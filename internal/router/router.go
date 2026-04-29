package router

import (
	"github.com/Pokem0n2/gym-log/internal/handlers"
	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
)

func New(store *repository.UserStore) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	auth := handlers.NewAuthHandler(store)
	exh := handlers.NewExerciseHandler()
	wh := handlers.NewWorkoutHandler()
	sh := handlers.NewSetHandler()
	sth := handlers.NewStatsHandler()

	// 公开路由
	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", auth.Login)
		api.POST("/auth/logout", auth.Logout)
	}

	// 需认证的路由
	authAPI := api.Group("", middleware.AuthRequired(store))
	{
		authAPI.GET("/auth/me", auth.Me)
		authAPI.POST("/auth/change-password", auth.ChangePassword)

		// 动作库
		authAPI.GET("/exercises", exh.List)
		authAPI.POST("/exercises", exh.Create)
		authAPI.DELETE("/exercises/:id", exh.Delete)

		// 训练记录
		authAPI.GET("/workouts", wh.List)
		authAPI.GET("/workouts/:id", wh.Get)
		authAPI.POST("/workouts", wh.Create)
		authAPI.DELETE("/workouts/:id", wh.Delete)

		// 组记录
		authAPI.POST("/workouts/:workout_id/sets", sh.Create)
		authAPI.DELETE("/sets/:id", sh.Delete)

		// 统计
		authAPI.GET("/stats/exercise/:exercise_id", sth.ExerciseHistory)
		authAPI.GET("/stats/volume", sth.VolumeByDate)
	}

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	r.GET("/login", func(c *gin.Context) {
		c.File("./static/login.html")
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
