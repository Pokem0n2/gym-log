package models

import "time"

// Exercise 动作定义
type Exercise struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// Workout 训练记录
type Workout struct {
	ID        int64     `json:"id"`
	Date      string    `json:"date" binding:"required"`
	Notes     string    `json:"notes"`
	Sets      []Set     `json:"sets,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Set 单组记录
type Set struct {
	ID         int64     `json:"id"`
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id" binding:"required"`
	Reps       int       `json:"reps" binding:"required,min=1"`
	Weight     float64   `json:"weight" binding:"required,min=0"`
	RPE        *float64  `json:"rpe,omitempty"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}
