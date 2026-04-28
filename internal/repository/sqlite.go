package repository

import (
	"database/sql"

	"github.com/Pokem0n2/gym-log/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func NewSQLite(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	if err := seed(db); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS exercises (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS workouts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS sets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
		exercise_id INTEGER NOT NULL REFERENCES exercises(id),
		reps INTEGER NOT NULL,
		weight REAL NOT NULL,
		rpe REAL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_sets_workout ON sets(workout_id);
	CREATE INDEX IF NOT EXISTS idx_workouts_date ON workouts(date);
	`
	_, err := db.Exec(schema)
	return err
}

// Exercise CRUD

func (db *DB) ListExercises() ([]models.Exercise, error) {
	rows, err := db.Query("SELECT id, name, category, created_at FROM exercises ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Exercise
	for rows.Next() {
		var e models.Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.Category, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (db *DB) CreateExercise(e *models.Exercise) error {
	r := db.QueryRow("INSERT INTO exercises(name, category) VALUES(?,?) RETURNING id, created_at", e.Name, e.Category)
	return r.Scan(&e.ID, &e.CreatedAt)
}

func (db *DB) DeleteExercise(id int64) error {
	_, err := db.Exec("DELETE FROM exercises WHERE id = ?", id)
	return err
}

// Workout CRUD

func (db *DB) ListWorkouts() ([]models.Workout, error) {
	rows, err := db.Query("SELECT id, date, notes, created_at FROM workouts ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Workout
	for rows.Next() {
		var w models.Workout
		if err := rows.Scan(&w.ID, &w.Date, &w.Notes, &w.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

func (db *DB) GetWorkout(id int64) (*models.Workout, error) {
	var w models.Workout
	if err := db.QueryRow("SELECT id, date, notes, created_at FROM workouts WHERE id = ?", id).Scan(&w.ID, &w.Date, &w.Notes, &w.CreatedAt); err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT s.id, s.workout_id, s.exercise_id, s.reps, s.weight, s.rpe, s.notes, s.created_at, e.name as exercise_name
		FROM sets s JOIN exercises e ON s.exercise_id = e.id
		WHERE s.workout_id = ? ORDER BY s.created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Set
		var ename string
		if err := rows.Scan(&s.ID, &s.WorkoutID, &s.ExerciseID, &s.Reps, &s.Weight, &s.RPE, &s.Notes, &s.CreatedAt, &ename); err != nil {
			return nil, err
		}
		w.Sets = append(w.Sets, s)
	}
	return &w, rows.Err()
}

func (db *DB) CreateWorkout(w *models.Workout) error {
	r := db.QueryRow("INSERT INTO workouts(date, notes) VALUES(?,?) RETURNING id, created_at", w.Date, w.Notes)
	return r.Scan(&w.ID, &w.CreatedAt)
}

func (db *DB) DeleteWorkout(id int64) error {
	_, err := db.Exec("DELETE FROM workouts WHERE id = ?", id)
	return err
}

// Set

func (db *DB) AddSet(s *models.Set) error {
	r := db.QueryRow("INSERT INTO sets(workout_id, exercise_id, reps, weight, rpe, notes) VALUES(?,?,?,?,?,?) RETURNING id, created_at",
		s.WorkoutID, s.ExerciseID, s.Reps, s.Weight, s.RPE, s.Notes)
	return r.Scan(&s.ID, &s.CreatedAt)
}

func (db *DB) DeleteSet(id int64) error {
	_, err := db.Exec("DELETE FROM sets WHERE id = ?", id)
	return err
}

// Stats

func (db *DB) GetExerciseStats(exerciseID int64) ([]models.Set, error) {
	rows, err := db.Query(`
		SELECT id, workout_id, exercise_id, reps, weight, rpe, notes, created_at
		FROM sets WHERE exercise_id = ? ORDER BY created_at DESC LIMIT 50`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Set
	for rows.Next() {
		var s models.Set
		if err := rows.Scan(&s.ID, &s.WorkoutID, &s.ExerciseID, &s.Reps, &s.Weight, &s.RPE, &s.Notes, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (db *DB) GetVolumeByDate(start, end string) (map[string]float64, error) {
	rows, err := db.Query(`
		SELECT w.date, SUM(s.weight * s.reps) as volume
		FROM sets s JOIN workouts w ON s.workout_id = w.id
		WHERE w.date BETWEEN ? AND ?
		GROUP BY w.date ORDER BY w.date`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var date string
		var vol float64
		if err := rows.Scan(&date, &vol); err != nil {
			return nil, err
		}
		result[date] = vol
	}
	return result, rows.Err()
}

func seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	exercises := []struct{ name, category string }{
		// 胸部
		{"杠铃卧推", "胸部"},
		{"哑铃卧推", "胸部"},
		{"上斜杠铃卧推", "胸部"},
		{"上斜哑铃卧推", "胸部"},
		{"哑铃飞鸟", "胸部"},
		{"绳索夾胸", "胸部"},
		{"俯卧撑", "胸部"},
		{"双杠臂曲伸", "胸部"},
		// 背部
		{"引体向上", "背部"},
		{"杠铃划船", "背部"},
		{"哑铃单臂划船", "背部"},
		{"高位下拉", "背部"},
		{"坐姿划船", "背部"},
		{"硬拉", "背部"},
		{"直腿硬拉", "背部"},
		{"反向飞鸟", "背部"},
		{"山羊挺身", "背部"},
		// 肩部
		{"杠铃推举", "肩部"},
		{"哑铃推举", "肩部"},
		{"侧平举", "肩部"},
		{"前平举", "肩部"},
		{"俯身飞鸟", "肩部"},
		{"面拉", "肩部"},
		{"杠铃耸肩", "肩部"},
		// 二头
		{"杠铃弯举", "二头肌"},
		{"哑铃弯举", "二头肌"},
		{"锤式弯举", "二头肌"},
		{"牧师凳弯举", "二头肌"},
		{"集中弯举", "二头肌"},
		// 三头
		{"绳索下压", "三头肌"},
		{"仰卧臂曲伸", "三头肌"},
		{"窄距卧推", "三头肌"},
		{"哑铃颈后臂曲伸", "三头肌"},
		{"俯身臂曲伸", "三头肌"},
		// 胯四头肌
		{"深蹲", "胯四头肌"},
		{"前蹲", "胯四头肌"},
		{"腿举", "胯四头肌"},
		{"腿伸屈", "胯四头肌"},
		{"弓步蹲", "胯四头肌"},
		{"保加利亚分腿蹲", "胯四头肌"},
		// 膘绳肌
		{"腿弯举", "膘绳肌"},
		{"早安式", "膘绳肌"},
		// 臀部
		{"臀推", "臀部"},
		{"壶铃摇摆", "臀部"},
		{"绳索后踢腿", "臀部"},
		// 核心
		{"卷腹", "核心"},
		{"悬垂举腿", "核心"},
		{"平板支撑", "核心"},
		{"俄罗斯转体", "核心"},
		{"仰卧抬腿", "核心"},
		// 有氧
		{"跑步机", "有氧"},
		{"椭圆机", "有氧"},
		{"划船机", "有氧"},
		{"战绳", "有氧"},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO exercises(name, category) VALUES(?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, e := range exercises {
		if _, err := stmt.Exec(e.name, e.category); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
