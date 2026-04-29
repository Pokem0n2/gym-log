package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Pokem0n2/gym-log/internal/models"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// UserStore 管理用户账号数据库和各自业务数据库
type UserStore struct {
	baseDir    string
	userDB     *sql.DB
	userDBPath string

	mu      sync.RWMutex
	dbCache map[string]*DB
}

// NewUserStore 初始化用户管理器
// baseDir: 用户数据根目录，如 "./data"
func NewUserStore(baseDir string) (*UserStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	userDBPath := filepath.Join(baseDir, "users.db")
	userDB, err := sql.Open("sqlite", userDBPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if err := userDB.Ping(); err != nil {
		return nil, err
	}
	if err := migrateUserDB(userDB); err != nil {
		return nil, err
	}

	store := &UserStore{
		baseDir:    baseDir,
		userDB:     userDB,
		userDBPath: userDBPath,
		dbCache:    make(map[string]*DB),
	}
	return store, nil
}

func migrateUserDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		user_id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(schema)
	return err
}

// Close 关闭所有数据库连接
func (s *UserStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, db := range s.dbCache {
		db.Close()
	}
	s.dbCache = nil
	return s.userDB.Close()
}

// CreateUser 管理员创建账号
func (s *UserStore) CreateUser(userID, username, plainPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.userDB.Exec(
		"INSERT INTO users(user_id, username, password_hash) VALUES(?,?,?)",
		userID, username, string(hash),
	)
	return err
}

// ValidateUser 验证登录
func (s *UserStore) ValidateUser(username, plainPassword string) (*models.User, error) {
	var u models.User
	var hash string
	err := s.userDB.QueryRow(
		"SELECT user_id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.UserID, &u.Username, &hash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	return &u, nil
}

// ChangePassword 修改密码
func (s *UserStore) ChangePassword(userID, oldPassword, newPassword string) error {
	var hash string
	err := s.userDB.QueryRow(
		"SELECT password_hash FROM users WHERE user_id = ?", userID,
	).Scan(&hash)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("原密码错误")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.userDB.Exec(
		"UPDATE users SET password_hash = ? WHERE user_id = ?",
		string(newHash), userID,
	)
	return err
}

// GetUserDB 获取指定用户的业务数据库（带缓存）
func (s *UserStore) GetUserDB(userID string) (*DB, error) {
	s.mu.RLock()
	if db, ok := s.dbCache[userID]; ok {
		s.mu.RUnlock()
		return db, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双检，防止并发重复创建
	if db, ok := s.dbCache[userID]; ok {
		return db, nil
	}

	userDir := filepath.Join(s.baseDir, "users", userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(userDir, "gym.db")
	db, err := NewSQLite(dbPath)
	if err != nil {
		return nil, err
	}
	s.dbCache[userID] = db
	return db, nil
}
