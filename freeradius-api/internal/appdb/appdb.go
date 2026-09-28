// Package appdb is a small SQLite store for application data that does NOT live
// in the FreeRADIUS MySQL database: admin login users, login history, and an
// activity log. Schema creation and seeding are idempotent, so every server
// start is safe to repeat.
package appdb

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS app_users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	role          TEXT NOT NULL DEFAULT 'admin',
	created_at    TEXT NOT NULL DEFAULT (datetime('now')),
	last_login_at TEXT
);
CREATE TABLE IF NOT EXISTS login_history (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER REFERENCES app_users(id) ON DELETE SET NULL,
	username   TEXT NOT NULL,
	ip         TEXT,
	user_agent TEXT,
	success    INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS activity_logs (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER REFERENCES app_users(id) ON DELETE SET NULL,
	username   TEXT,
	action     TEXT NOT NULL,
	method     TEXT,
	path       TEXT,
	status     INTEGER,
	ip         TEXT,
	user_agent TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_login_history_created  ON login_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_username ON login_history(username);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created  ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_username ON activity_logs(username);
`

// Open opens (creating if needed) the SQLite file, applies the schema and the
// admin seed. Safe to call once per process; repeated calls/restarts are no-ops.
func Open(path string) error {
	dsn := "file:" + path + "?_busy_timeout=5000&_journal_mode=WAL&_fk=1"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	// ponytail: single writer connection — SQLite serializes writes anyway and
	// this avoids "database is locked". Bump if logging ever becomes hot.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return err
	}
	DB = db
	return nil
}

func Close() {
	if DB != nil {
		_ = DB.Close()
	}
}

// SeedAdmin inserts the configured admin once. INSERT OR IGNORE keeps existing
// rows (and changed passwords) intact across restarts.
func SeedAdmin(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec(
		`INSERT OR IGNORE INTO app_users (username, password_hash, role) VALUES (?, ?, 'admin')`,
		username, string(hash))
	return err
}

// Authenticate checks username/password against the stored bcrypt hash.
func Authenticate(username, password string) (id int64, role string, ok bool) {
	var hash string
	err := DB.QueryRow(`SELECT id, role, password_hash FROM app_users WHERE username = ?`, username).
		Scan(&id, &role, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return 0, "", false
	}
	return id, role, true
}

// UserID returns the id for a username (0 when unknown).
func UserID(username string) int64 {
	var id int64
	_ = DB.QueryRow(`SELECT id FROM app_users WHERE username = ?`, username).Scan(&id)
	return id
}

// RecordLogin appends a login attempt and, on success, stamps last_login_at.
func RecordLogin(userID int64, username, ip, ua string, success bool) error {
	succ := 0
	if success {
		succ = 1
	}
	var uid interface{}
	if userID > 0 {
		uid = userID
	}
	if _, err := DB.Exec(
		`INSERT INTO login_history (user_id, username, ip, user_agent, success) VALUES (?, ?, ?, ?, ?)`,
		uid, username, ip, ua, succ); err != nil {
		return err
	}
	if success && userID > 0 {
		_, _ = DB.Exec(`UPDATE app_users SET last_login_at = datetime('now') WHERE id = ?`, userID)
	}
	return nil
}

// RecordActivity appends an activity (audit) entry.
func RecordActivity(userID int64, username, action, method, path string, status int, ip, ua string) error {
	var uid interface{}
	if userID > 0 {
		uid = userID
	}
	var uname interface{}
	if username != "" {
		uname = username
	}
	_, err := DB.Exec(
		`INSERT INTO activity_logs (user_id, username, action, method, path, status, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uid, uname, action, method, path, status, ip, ua)
	return err
}

// ---- read side (for the dashboard) ----

type LoginRecord struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Success   bool   `json:"success"`
	CreatedAt string `json:"created_at"`
}

type ActivityRecord struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	IP        string `json:"ip"`
	CreatedAt string `json:"created_at"`
}

type UserRecord struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
	LastLoginAt string `json:"last_login_at"`
}

func RecentLogins(limit int) ([]LoginRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := DB.Query(
		`SELECT id, username, COALESCE(ip,''), COALESCE(user_agent,''), success, created_at
		 FROM login_history ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LoginRecord{}
	for rows.Next() {
		var r LoginRecord
		var succ int
		if err := rows.Scan(&r.ID, &r.Username, &r.IP, &r.UserAgent, &succ, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Success = succ == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

func RecentActivity(limit int) ([]ActivityRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := DB.Query(
		`SELECT id, COALESCE(username,''), action, COALESCE(method,''), COALESCE(path,''),
		        COALESCE(status,0), COALESCE(ip,''), created_at
		 FROM activity_logs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ActivityRecord{}
	for rows.Next() {
		var r ActivityRecord
		if err := rows.Scan(&r.ID, &r.Username, &r.Action, &r.Method, &r.Path, &r.Status, &r.IP, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func ListUsers() ([]UserRecord, error) {
	rows, err := DB.Query(
		`SELECT id, username, role, created_at, COALESCE(last_login_at,'')
		 FROM app_users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UserRecord{}
	for rows.Next() {
		var u UserRecord
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.LastLoginAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// LastSuccessfulLogin returns the most recent successful login time before now.
func LastSuccessfulLogin(username string) (string, bool) {
	var t string
	err := DB.QueryRow(
		`SELECT created_at FROM login_history WHERE username = ? AND success = 1
		 ORDER BY id DESC LIMIT 1`, username).Scan(&t)
	return t, err == nil
}

// Stats returns simple counters for the dashboard.
func Stats() map[string]int64 {
	m := map[string]int64{}
	var users, logins, failed, activity int64
	_ = DB.QueryRow(`SELECT COUNT(*) FROM app_users`).Scan(&users)
	_ = DB.QueryRow(`SELECT COUNT(*) FROM login_history`).Scan(&logins)
	_ = DB.QueryRow(`SELECT COUNT(*) FROM login_history WHERE success = 0`).Scan(&failed)
	_ = DB.QueryRow(`SELECT COUNT(*) FROM activity_logs`).Scan(&activity)
	m["users"] = users
	m["logins"] = logins
	m["failed_logins"] = failed
	m["activity"] = activity
	return m
}
