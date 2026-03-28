package models

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vinovest/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int       `db:"id"`
	Name           string    `db:"name"`
	Email          string    `db:"email"`
	HashedPassword []byte    `db:"hashed_password"`
	Created        time.Time `db:"created"`
}

type UserModel struct {
	DB *sqlx.DB
}

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created)
             VALUES ($1, $2, $3, now())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// 23505 is the Postgres code for unique_violation
			if pgErr.Code == "23505" &&
				pgErr.ConstraintName == "users_uc_email" {
				log.Print(err)
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

// Authenticate method to verify whether a user exists with the
// provided email address and password. returns the relevant user ID
// if they do.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	return id, nil
}

// Exists method to check if a user exists with a specific ID.
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
