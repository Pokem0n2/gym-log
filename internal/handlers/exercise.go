package handlers

import (
	"net/http"
	"strconv"

	"github.com/Pokem0n2/gym-log/internal/models"
	"github.com/Pokem0n2/gym-log/internal/repository"
	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	db *repository.DB
}

func NewExerciseHandler(db *repository.DB) *ExerciseHandler {
	return &ExerciseHandler{db: db}
}

func (h *ExerciseHandler) List(c *gin.Context) {
	list, err := h.db.ListExercises()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ExerciseHandler) Create(c *gin.Context) {
	var e models.Exercise
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.CreateExercise(&e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *ExerciseHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.DeleteExercise(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
