package models

import "time"

// Exercise 动作定义
type Exercise struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Category  string    `json:"category"`
	Fields    string    `json:"fields"`
	CreatedAt time.Time `json:"created_at"`
}

// Workout 训练记录
type Workout struct {
	ID         int64     `json:"id"`
	Date       string    `json:"date" binding:"required"`
	Notes      string    `json:"notes"`
	Sets       []Set     `json:"sets,omitempty"`
	TimeRanges []string  `json:"time_ranges,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Set 单组记录
type Set struct {
	ID           int64     `json:"id"`
	WorkoutID    int64     `json:"workout_id"`
	ExerciseID   int64     `json:"exercise_id" binding:"required"`
	ExerciseName string    `json:"exercise_name"`
	Reps         int       `json:"reps"`
	Weight       float64   `json:"weight"`
	RPE          *float64  `json:"rpe,omitempty"`
	IsWarmup     bool      `json:"is_warmup"`
	Extra        string    `json:"extra"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
}
