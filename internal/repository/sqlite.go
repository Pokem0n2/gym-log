package repository

import (
	"database/sql"
	"time"

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
		is_warmup INTEGER DEFAULT 0,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_sets_workout ON sets(workout_id);
	CREATE INDEX IF NOT EXISTS idx_workouts_date ON workouts(date);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	// 兼容旧表：尝试添加新列
	_, _ = db.Exec("ALTER TABLE exercises ADD COLUMN fields TEXT DEFAULT '[\\\"weight\\\",\\\"reps\\\",\\\"sets\\\"]'")
	_, _ = db.Exec("ALTER TABLE sets ADD COLUMN extra TEXT DEFAULT '{}'")
	_, _ = db.Exec("ALTER TABLE sets ADD COLUMN is_warmup INTEGER DEFAULT 0")
	// 旧数据迁移：为已有动作补充默认 fields
	return migrateExerciseFields(db)
}

func migrateExerciseFields(db *sql.DB) error {
	// 无条件更新所有预设动作的 fields，确保与代码中的定义一致
	for _, e := range presetExercises {
		_, _ = db.Exec("UPDATE exercises SET fields = ? WHERE name = ?", e.fields, e.name)
	}
	// 其余未设置的动作，默认为力量型
	_, _ = db.Exec("UPDATE exercises SET fields = '[\"weight\",\"reps\",\"sets\"]' WHERE fields IS NULL OR fields = ''")
	return nil
}

// 预设动作列表（seed 和迁移共用）
var presetExercises = []struct{ name, category, fields string }{
	// 胸部
	{"杠铃卧推", "胸部", `["weight","reps","sets"]`},
	{"哑铃卧推", "胸部", `["weight","reps","sets"]`},
	{"上斜杠铃卧推", "胸部", `["weight","reps","sets"]`},
	{"上斜哑铃卧推", "胸部", `["weight","reps","sets"]`},
	{"哑铃飞鸟", "胸部", `["weight","reps","sets"]`},
	{"绳索夹胸", "胸部", `["weight","reps","sets"]`},
	{"俯卧撑", "胸部", `["reps","sets"]`},
	{"双杠臂曲伸", "胸部", `["reps","sets"]`},
	// 背部
	{"引体向上", "背部", `["reps","sets"]`},
	{"杠铃划船", "背部", `["weight","reps","sets"]`},
	{"哑铃单臂划船", "背部", `["weight","reps","sets"]`},
	{"高位下拉", "背部", `["weight","reps","sets"]`},
	{"坐姿划船", "背部", `["weight","reps","sets"]`},
	{"硬拉", "背部", `["weight","reps","sets"]`},
	{"直腿硬拉", "背部", `["weight","reps","sets"]`},
	{"反向飞鸟", "背部", `["weight","reps","sets"]`},
	{"山羊挺身", "背部", `["weight","reps","sets"]`},
	// 肩部
	{"杠铃推举", "肩部", `["weight","reps","sets"]`},
	{"哑铃推举", "肩部", `["weight","reps","sets"]`},
	{"侧平举", "肩部", `["weight","reps","sets"]`},
	{"前平举", "肩部", `["weight","reps","sets"]`},
	{"俯身飞鸟", "肩部", `["weight","reps","sets"]`},
	{"面拉", "肩部", `["weight","reps","sets"]`},
	{"杠铃耸肩", "肩部", `["weight","reps","sets"]`},
	// 二头肌
	{"杠铃弯举", "二头肌", `["weight","reps","sets"]`},
	{"哑铃弯举", "二头肌", `["weight","reps","sets"]`},
	{"锤式弯举", "二头肌", `["weight","reps","sets"]`},
	{"牧师凳弯举", "二头肌", `["weight","reps","sets"]`},
	{"集中弯举", "二头肌", `["weight","reps","sets"]`},
	// 三头肌
	{"绳索下压", "三头肌", `["weight","reps","sets"]`},
	{"仰卧臂曲伸", "三头肌", `["weight","reps","sets"]`},
	{"窄距卧推", "三头肌", `["weight","reps","sets"]`},
	{"哑铃颈后臂曲伸", "三头肌", `["weight","reps","sets"]`},
	{"俯身臂曲伸", "三头肌", `["weight","reps","sets"]`},
	// 胯四头肌
	{"深蹲", "胯四头肌", `["weight","reps","sets"]`},
	{"前蹲", "胯四头肌", `["weight","reps","sets"]`},
	{"腿举", "胯四头肌", `["weight","reps","sets"]`},
	{"腿伸屈", "胯四头肌", `["weight","reps","sets"]`},
	{"弓步蹲", "胯四头肌", `["weight","reps","sets"]`},
	{"保加利亚分腿蹲", "胯四头肌", `["weight","reps","sets"]`},
	// 膘绳肌
	{"腿弯举", "膘绳肌", `["weight","reps","sets"]`},
	{"早安式", "膘绳肌", `["weight","reps","sets"]`},
	// 臀部
	{"臀推", "臀部", `["weight","reps","sets"]`},
	{"壶铃摆摇", "臀部", `["weight","reps","sets"]`},
	{"绳索后踢腿", "臀部", `["weight","reps","sets"]`},
	// 核心
	{"卷腹", "核心", `["reps","sets"]`},
	{"悬垂举腿", "核心", `["reps","sets"]`},
	{"平板支撑", "核心", `["duration","sets"]`},
	{"俄罗斯转体", "核心", `["reps","sets"]`},
	{"仰卧抬腿", "核心", `["reps","sets"]`},
	// 有氧
	{"跑步机", "有氧", `["speed","incline","duration"]`},
	{"椭圆机", "有氧", `["resistance","incline","duration"]`},
	{"划船机", "有氧", `["distance","duration"]`},
	{"战绳", "有氧", `["duration","sets"]`},
}

func seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	exercises := presetExercises

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO exercises(name, category, fields) VALUES(?,?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, e := range exercises {
		if _, err := stmt.Exec(e.name, e.category, e.fields); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Exercise CRUD

func (db *DB) ListExercises() ([]models.Exercise, error) {
	rows, err := db.Query("SELECT id, name, category, fields, created_at FROM exercises ORDER BY category, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Exercise
	for rows.Next() {
		var e models.Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.Category, &e.Fields, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (db *DB) CreateExercise(e *models.Exercise) error {
	if e.Fields == "" {
		e.Fields = `["weight","reps","sets"]`
	}
	r := db.QueryRow("INSERT INTO exercises(name, category, fields) VALUES(?,?,?) RETURNING id, created_at", e.Name, e.Category, e.Fields)
	return r.Scan(&e.ID, &e.CreatedAt)
}

func (db *DB) DeleteExercise(id int64) error {
	_, err := db.Exec("DELETE FROM exercises WHERE id = ?", id)
	return err
}

// Workout CRUD

func (db *DB) ListWorkouts() ([]models.Workout, error) {
	rows, err := db.Query(`
		SELECT w.id, w.date, w.notes, w.created_at,
			s.id, s.workout_id, s.exercise_id, s.reps, s.weight, s.rpe, s.is_warmup, s.extra, s.notes, s.created_at, e.name as exercise_name
		FROM workouts w
		LEFT JOIN sets s ON w.id = s.workout_id
		LEFT JOIN exercises e ON s.exercise_id = e.id
		ORDER BY w.date DESC, w.created_at DESC, s.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Workout
	var current *models.Workout
	for rows.Next() {
		var w models.Workout
		var s models.Set
		var sid sql.NullInt64
		var sWorkoutID sql.NullInt64
		var sExerciseID sql.NullInt64
		var sReps sql.NullInt64
		var sWeight sql.NullFloat64
		var sRPE sql.NullFloat64
		var sIsWarmup sql.NullInt64
		var sExtra sql.NullString
		var sNotes sql.NullString
		var sCreatedAt sql.NullTime
		var sExName sql.NullString

		if err := rows.Scan(&w.ID, &w.Date, &w.Notes, &w.CreatedAt,
			&sid, &sWorkoutID, &sExerciseID, &sReps, &sWeight, &sRPE, &sIsWarmup, &sExtra, &sNotes, &sCreatedAt, &sExName); err != nil {
			return nil, err
		}

		if current == nil || current.ID != w.ID {
			if current != nil {
				list = append(list, *current)
			}
			current = &w
		}

		if sid.Valid {
			s.ID = sid.Int64
			s.WorkoutID = sWorkoutID.Int64
			s.ExerciseID = sExerciseID.Int64
			s.Reps = int(sReps.Int64)
			s.Weight = sWeight.Float64
			if sRPE.Valid {
				rpe := sRPE.Float64
				s.RPE = &rpe
			}
			s.IsWarmup = sIsWarmup.Int64 != 0
			s.Extra = sExtra.String
			s.Notes = sNotes.String
			s.CreatedAt = sCreatedAt.Time
			s.ExerciseName = sExName.String
			current.Sets = append(current.Sets, s)
		}
	}
	if current != nil {
		list = append(list, *current)
	}
	return list, rows.Err()
}

func (db *DB) GetWorkout(id int64) (*models.Workout, error) {
	var w models.Workout
	if err := db.QueryRow("SELECT id, date, notes, created_at FROM workouts WHERE id = ?", id).Scan(&w.ID, &w.Date, &w.Notes, &w.CreatedAt); err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT s.id, s.workout_id, s.exercise_id, s.reps, s.weight, s.rpe, s.is_warmup, s.extra, s.notes, s.created_at, e.name as exercise_name
		FROM sets s JOIN exercises e ON s.exercise_id = e.id
		WHERE s.workout_id = ? ORDER BY s.created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Set
		if err := rows.Scan(&s.ID, &s.WorkoutID, &s.ExerciseID, &s.Reps, &s.Weight, &s.RPE, &s.IsWarmup, &s.Extra, &s.Notes, &s.CreatedAt, &s.ExerciseName); err != nil {
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
	r := db.QueryRow("INSERT INTO sets(workout_id, exercise_id, reps, weight, rpe, is_warmup, extra, notes) VALUES(?,?,?,?,?,?,?,?) RETURNING id, created_at",
		s.WorkoutID, s.ExerciseID, s.Reps, s.Weight, s.RPE, s.IsWarmup, s.Extra, s.Notes)
	return r.Scan(&s.ID, &s.CreatedAt)
}

func (db *DB) DeleteSet(id int64) error {
	_, err := db.Exec("DELETE FROM sets WHERE id = ?", id)
	return err
}

// Stats

func (db *DB) GetExerciseStats(exerciseID int64) ([]models.Set, error) {
	rows, err := db.Query(`
		SELECT s.id, s.workout_id, s.exercise_id, s.reps, s.weight, s.rpe, s.is_warmup, s.extra, s.notes, s.created_at, w.date as workout_date
		FROM sets s JOIN workouts w ON s.workout_id = w.id
		WHERE s.exercise_id = ? AND s.is_warmup = 0 ORDER BY w.date DESC, s.created_at DESC LIMIT 50`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Set
	for rows.Next() {
		var s models.Set
		if err := rows.Scan(&s.ID, &s.WorkoutID, &s.ExerciseID, &s.Reps, &s.Weight, &s.RPE, &s.IsWarmup, &s.Extra, &s.Notes, &s.CreatedAt, &s.WorkoutDate); err != nil {
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
		WHERE w.date BETWEEN ? AND ? AND s.is_warmup = 0
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

func (db *DB) GetWorkoutTimeRanges(workoutID int64) ([]string, error) {
	rows, err := db.Query("SELECT created_at FROM sets WHERE workout_id = ? ORDER BY created_at", workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var times []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		times = append(times, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(times) == 0 {
		return nil, nil
	}

	const twoHours = 2 * time.Hour
	var ranges []string
	start := times[0]
	end := times[0]
	loc := time.FixedZone("CST", 8*3600)

	for i := 1; i < len(times); i++ {
		if times[i].Sub(end) <= twoHours {
			end = times[i]
		} else {
			ranges = append(ranges, formatTimeRange(start, end, loc))
			start = times[i]
			end = times[i]
		}
	}
	ranges = append(ranges, formatTimeRange(start, end, loc))
	return ranges, nil
}

func formatTimeRange(start, end time.Time, loc *time.Location) string {
	s := start.In(loc).Format("15:04")
	e := end.In(loc).Format("15:04")
	if s == e {
		return s
	}
	return s + "~" + e
}
