package handlers

import (
	"net/http"
	"strconv"

	"github.com/Pokem0n2/gym-log/internal/models"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
)

type WorkoutHandler struct {
	db *repository.DB
}

func NewWorkoutHandler(db *repository.DB) *WorkoutHandler {
	return &WorkoutHandler{db: db}
}

func (h *WorkoutHandler) List(c *gin.Context) {
	list, err := h.db.ListWorkouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *WorkoutHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	w, err := h.db.GetWorkout(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WorkoutHandler) Create(c *gin.Context) {
	var w models.Workout
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.CreateWorkout(&w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WorkoutHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.DeleteWorkout(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// SetHandler 处理单组记录

type SetHandler struct {
	db *repository.DB
}

func NewSetHandler(db *repository.DB) *SetHandler {
	return &SetHandler{db: db}
}

func (h *SetHandler) Create(c *gin.Context) {
	wid, _ := strconv.ParseInt(c.Param("workout_id"), 10, 64)
	var s models.Set
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.WorkoutID = wid
	if err := h.db.AddSet(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *SetHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.DeleteSet(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// StatsHandler

type StatsHandler struct {
	db *repository.DB
}

func NewStatsHandler(db *repository.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

func (h *StatsHandler) ExerciseHistory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("exercise_id"), 10, 64)
	sets, err := h.db.GetExerciseStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sets)
}

func (h *StatsHandler) VolumeByDate(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")
	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required"})
		return
	}
	data, err := h.db.GetVolumeByDate(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
