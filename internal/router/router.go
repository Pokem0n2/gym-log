package router

import (
	"github.com/Pokem0n2/gym-log/internal/handlers"
	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
)

func New(db *repository.DB) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	exh := handlers.NewExerciseHandler(db)
	wh := handlers.NewWorkoutHandler(db)
	sh := handlers.NewSetHandler(db)
	sth := handlers.NewStatsHandler(db)

	api := r.Group("/api/v1")
	{
		// 动作库
		api.GET("/exercises", exh.List)
		api.POST("/exercises", exh.Create)
		api.DELETE("/exercises/:id", exh.Delete)

		// 训练记录
		api.GET("/workouts", wh.List)
		api.GET("/workouts/:id", wh.Get)
		api.POST("/workouts", wh.Create)
		api.DELETE("/workouts/:id", wh.Delete)

		// 组记录
		api.POST("/workouts/:workout_id/sets", sh.Create)
		api.DELETE("/sets/:id", sh.Delete)

		// 统计
		api.GET("/stats/exercise/:exercise_id", sth.ExerciseHistory)
		api.GET("/stats/volume", sth.VolumeByDate)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
