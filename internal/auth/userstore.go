package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// UserInfo is the safe-to-return user representation (no secrets)
type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type userRecord struct {
	username string
	hash     string
	secret   string
	role     string
}

// UserStore manages users in SQLite (app_user for admin id=1, app_sub_users for others)
type UserStore struct {
	mu   sync.Mutex
	db   *sql.DB
	dir  string
	done bool
}

type legacyUserFile struct {
	Username      string `json:"username"`
	PasswordHash  string `json:"passwordHash"`
	SessionSecret string `json:"sessionSecret"`
}

// LoadUserStore loads the user store and runs a one-time JSON migration if needed.
func LoadUserStore(db *sql.DB, dataDir string) (*UserStore, error) {
	u := &UserStore{db: db, dir: dataDir}
	if err := u.migrateFromJSONIfNeeded(); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UserStore) migrateFromJSONIfNeeded() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.done {
		return nil
	}
	u.done = true
	var username string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE id=1`).Scan(&username)
	if username != "" {
		return nil
	}
	path := filepath.Join(u.dir, "users.json")
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		return nil
	}
	var leg legacyUserFile
	if json.Unmarshal(b, &leg) != nil || leg.Username == "" || leg.PasswordHash == "" {
		return nil
	}
	_, err = u.db.Exec(`UPDATE app_user SET username=?, password_hash=?, session_secret=? WHERE id=1`,
		leg.Username, leg.PasswordHash, leg.SessionSecret)
	return err
}

// findByUsernameLocked looks up a user in app_user then app_sub_users. Caller must hold the lock.
func (u *UserStore) findByUsernameLocked(username string) (*userRecord, error) {
	var rec userRecord
	err := u.db.QueryRow(
		`SELECT username, password_hash, session_secret, COALESCE(role,'admin') FROM app_user WHERE id=1 AND username=?`,
		username,
	).Scan(&rec.username, &rec.hash, &rec.secret, &rec.role)
	if err == nil && rec.username != "" {
		return &rec, nil
	}
	err = u.db.QueryRow(
		`SELECT username, password_hash, session_secret, role FROM app_sub_users WHERE username=?`,
		username,
	).Scan(&rec.username, &rec.hash, &rec.secret, &rec.role)
	if err == nil {
		return &rec, nil
	}
	return nil, errors.New("user not found")
}

// HasUser reports whether the admin account has been created.
func (u *UserStore) HasUser() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	var name string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE id=1`).Scan(&name)
	return name != ""
}

func (u *UserStore) loadLocked() (username, hash, secret string, err error) {
	err = u.db.QueryRow(`SELECT username, password_hash, session_secret FROM app_user WHERE id=1`).
		Scan(&username, &hash, &secret)
	return
}

// CreateFirstUser creates the admin account (only when no user exists yet).
func (u *UserStore) CreateFirstUser(username, password string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	var existing string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE id=1`).Scan(&existing)
	if existing != "" {
		return errors.New("user already exists")
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("invalid password: %w. %s", err, GetPasswordStrengthHint())
	}
	pwhash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	sec := make([]byte, 32)
	if _, err := rand.Read(sec); err != nil {
		return fmt.Errorf("failed to generate session secret: %w", err)
	}
	secret := hex.EncodeToString(sec)
	_, err = u.db.Exec(
		`UPDATE app_user SET username=?, password_hash=?, session_secret=?, role='admin' WHERE id=1`,
		username, string(pwhash), secret,
	)
	return err
}

// VerifyPassword checks credentials and returns the UserInfo if valid.
func (u *UserStore) VerifyPassword(username, password string) (*UserInfo, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	rec, err := u.findByUsernameLocked(username)
	if err != nil || rec.hash == "" {
		return nil, false
	}
	if bcrypt.CompareHashAndPassword([]byte(rec.hash), []byte(password)) != nil {
		return nil, false
	}
	return &UserInfo{Username: rec.username, Role: rec.role}, true
}

// SessionKeyFor returns the HMAC key for the given username.
func (u *UserStore) SessionKeyFor(username string) []byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	rec, err := u.findByUsernameLocked(username)
	if err != nil || rec.secret == "" {
		return nil
	}
	b, err := hex.DecodeString(rec.secret)
	if err != nil || len(b) == 0 {
		return nil
	}
	return b
}

// sessionKey returns the admin's HMAC key (kept for backward compat).
func (u *UserStore) sessionKey() []byte {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, _, secret, err := u.loadLocked()
	if err != nil || secret == "" {
		return nil
	}
	b, err := hex.DecodeString(secret)
	if err != nil || len(b) == 0 {
		return nil
	}
	return b
}

// Username returns the admin username.
func (u *UserStore) Username() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	var name string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE id=1`).Scan(&name)
	return name
}

// GetRole returns the role string for the given username.
func (u *UserStore) GetRole(username string) string {
	u.mu.Lock()
	defer u.mu.Unlock()
	rec, err := u.findByUsernameLocked(username)
	if err != nil {
		return ""
	}
	return rec.role
}

// ChangePasswordFor changes the password for any user, requiring the old password.
func (u *UserStore) ChangePasswordFor(username, oldPassword, newPassword string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	rec, err := u.findByUsernameLocked(username)
	if err != nil || rec.username == "" {
		return errors.New("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(rec.hash), []byte(oldPassword)) != nil {
		return errors.New("当前密码错误")
	}
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("invalid new password: %w. %s", err, GetPasswordStrengthHint())
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	sec := make([]byte, 32)
	if _, err := rand.Read(sec); err != nil {
		return fmt.Errorf("failed to generate session secret: %w", err)
	}
	secret := hex.EncodeToString(sec)
	if rec.role == "admin" {
		_, err = u.db.Exec(`UPDATE app_user SET password_hash=?, session_secret=? WHERE id=1`, string(newHash), secret)
	} else {
		_, err = u.db.Exec(`UPDATE app_sub_users SET password_hash=?, session_secret=? WHERE username=?`, string(newHash), secret, username)
	}
	return err
}

// ChangePassword is the legacy admin-only password change (backward compat).
func (u *UserStore) ChangePassword(oldPassword, newPassword string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	name, hash, _, err := u.loadLocked()
	if err != nil || name == "" {
		return errors.New("no user")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPassword)) != nil {
		return errors.New("当前密码错误")
	}
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("invalid new password: %w. %s", err, GetPasswordStrengthHint())
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	sec := make([]byte, 32)
	if _, err := rand.Read(sec); err != nil {
		return fmt.Errorf("failed to generate session secret: %w", err)
	}
	secret := hex.EncodeToString(sec)
	_, err = u.db.Exec(`UPDATE app_user SET password_hash=?, session_secret=? WHERE id=1`, string(newHash), secret)
	return err
}

// ListUsers returns all users (admin first, then sub-users).
func (u *UserStore) ListUsers() ([]UserInfo, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	var users []UserInfo
	var name, role string
	if err := u.db.QueryRow(`SELECT username, COALESCE(role,'admin') FROM app_user WHERE id=1`).Scan(&name, &role); err == nil && name != "" {
		users = append(users, UserInfo{Username: name, Role: role})
	}
	rows, err := u.db.Query(`SELECT username, role FROM app_sub_users ORDER BY id`)
	if err != nil {
		return users, nil
	}
	defer rows.Close()
	for rows.Next() {
		var info UserInfo
		if err := rows.Scan(&info.Username, &info.Role); err != nil {
			continue
		}
		users = append(users, info)
	}
	return users, nil
}

// CreateUser creates a new sub-user. Role must be "user1" or "user2".
func (u *UserStore) CreateUser(username, password, role string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if role != "user1" && role != "user2" {
		return errors.New("role must be user1 or user2")
	}
	var existing string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE username=?`, username).Scan(&existing)
	if existing != "" {
		return errors.New("username already exists")
	}
	_ = u.db.QueryRow(`SELECT username FROM app_sub_users WHERE username=?`, username).Scan(&existing)
	if existing != "" {
		return errors.New("username already exists")
	}
	if err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("invalid password: %w. %s", err, GetPasswordStrengthHint())
	}
	pwhash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}
	sec := make([]byte, 32)
	if _, err := rand.Read(sec); err != nil {
		return err
	}
	secret := hex.EncodeToString(sec)
	_, err = u.db.Exec(
		`INSERT INTO app_sub_users (username, password_hash, session_secret, role) VALUES (?,?,?,?)`,
		username, string(pwhash), secret, role,
	)
	return err
}

// DeleteUser deletes a sub-user. Cannot delete the admin.
func (u *UserStore) DeleteUser(username string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	var adminName string
	_ = u.db.QueryRow(`SELECT username FROM app_user WHERE id=1`).Scan(&adminName)
	if username == adminName {
		return errors.New("cannot delete admin user")
	}
	result, err := u.db.Exec(`DELETE FROM app_sub_users WHERE username=?`, username)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}

// SetUserRole changes a sub-user's role. Cannot change admin's role.
func (u *UserStore) SetUserRole(username, role string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if role != "user1" && role != "user2" {
		return errors.New("role must be user1 or user2")
	}
	result, err := u.db.Exec(`UPDATE app_sub_users SET role=? WHERE username=?`, role, username)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New("user not found or is admin")
	}
	return nil
}

// AdminResetPassword resets any user's password without requiring the old password.
func (u *UserStore) AdminResetPassword(username, newPassword string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	rec, err := u.findByUsernameLocked(username)
	if err != nil {
		return errors.New("user not found")
	}
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("invalid password: %w. %s", err, GetPasswordStrengthHint())
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	sec := make([]byte, 32)
	if _, err := rand.Read(sec); err != nil {
		return err
	}
	secret := hex.EncodeToString(sec)
	if rec.role == "admin" {
		_, err = u.db.Exec(`UPDATE app_user SET password_hash=?, session_secret=? WHERE id=1`, string(newHash), secret)
	} else {
		_, err = u.db.Exec(`UPDATE app_sub_users SET password_hash=?, session_secret=? WHERE username=?`, string(newHash), secret, username)
	}
	return err
}
