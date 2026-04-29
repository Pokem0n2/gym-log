package handlers

import (
	"net/http"
	"strconv"

	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/models"
	"github.com/gin-gonic/gin"
)

type WorkoutHandler struct{}

func NewWorkoutHandler() *WorkoutHandler {
	return &WorkoutHandler{}
}

func (h *WorkoutHandler) List(c *gin.Context) {
	db := middleware.GetUserDB(c)
	list, err := db.ListWorkouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range list {
		ranges, _ := db.GetWorkoutTimeRanges(list[i].ID)
		list[i].TimeRanges = ranges
	}
	c.JSON(http.StatusOK, list)
}

func (h *WorkoutHandler) Get(c *gin.Context) {
	db := middleware.GetUserDB(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	w, err := db.GetWorkout(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WorkoutHandler) Create(c *gin.Context) {
	db := middleware.GetUserDB(c)
	var w models.Workout
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.CreateWorkout(&w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WorkoutHandler) Delete(c *gin.Context) {
	db := middleware.GetUserDB(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := db.DeleteWorkout(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// SetHandler 处理单组记录

type SetHandler struct{}

func NewSetHandler() *SetHandler {
	return &SetHandler{}
}

func (h *SetHandler) Create(c *gin.Context) {
	db := middleware.GetUserDB(c)
	wid, _ := strconv.ParseInt(c.Param("workout_id"), 10, 64)
	var s models.Set
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.WorkoutID = wid
	if err := db.AddSet(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *SetHandler) Delete(c *gin.Context) {
	db := middleware.GetUserDB(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := db.DeleteSet(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// StatsHandler

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

func (h *StatsHandler) ExerciseHistory(c *gin.Context) {
	db := middleware.GetUserDB(c)
	id, _ := strconv.ParseInt(c.Param("exercise_id"), 10, 64)
	sets, err := db.GetExerciseStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sets)
}

func (h *StatsHandler) VolumeByDate(c *gin.Context) {
	db := middleware.GetUserDB(c)
	start := c.Query("start")
	end := c.Query("end")
	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required"})
		return
	}
	data, err := db.GetVolumeByDate(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
