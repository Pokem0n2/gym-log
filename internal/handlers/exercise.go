package handlers

import (
	"net/http"
	"strconv"

	"github.com/Pokem0n2/gym-log/internal/middleware"
	"github.com/Pokem0n2/gym-log/internal/models"
	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct{}

func NewExerciseHandler() *ExerciseHandler {
	return &ExerciseHandler{}
}

func (h *ExerciseHandler) List(c *gin.Context) {
	db := middleware.GetUserDB(c)
	list, err := db.ListExercises()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ExerciseHandler) Create(c *gin.Context) {
	db := middleware.GetUserDB(c)
	var e models.Exercise
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.CreateExercise(&e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *ExerciseHandler) Delete(c *gin.Context) {
	db := middleware.GetUserDB(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := db.DeleteExercise(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
