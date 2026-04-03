package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vinovest/sqlx"
)

type Snippet struct {
	ID      int       `db:"id"`
	Title   string    `db:"title"`
	Content string    `db:"content"`
	Created time.Time `db:"created"`
	Expires time.Time `db:"expires"`
}

type SnippetModel struct {
	DB *sqlx.DB
}

type SnippetModelInterface interface {
	Insert(title, content string, expires int) (int, error)
	Get(id int) (Snippet, error)
	Latest() ([]Snippet, error)
	Delete(id int) error
}

// This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	// We can remove 'created' field as its defaults to now()
	stmt := `INSERT INTO snippets (title, content, created, expires)
             VALUES ($1, $2, now(), now() + make_interval(days => $3))
             RETURNING id`

	var id int
	// sqlx.Get is a wrapper for QueryRow + Scan
	err := m.DB.Get(&id, stmt, title, content, expires)
	if err != nil {
		return 0, fmt.Errorf("DB Insert: %w", err)
	}
	return id, nil
}

// Delete(id int) will delete the snippet with id.
func (m *SnippetModel) Delete(id int) error {
	stmt := `DELETE FROM snippets WHERE id = $1`
	result, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoRecord
	}
	return nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	// We filter by ID and also ensure the snippet hasn't expired yet
	stmt := `SELECT id, title, content, created, expires FROM snippets
             WHERE expires > now() AND id = $1`

	var s Snippet
	if err := m.DB.Get(&s, stmt, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}
	return s, nil
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
             WHERE expires > now() ORDER BY id DESC LIMIT 10;`

	var snippets []Snippet

	if err := m.DB.Select(&snippets, stmt); err != nil {
		return nil, fmt.Errorf("DB Latest: %w", err)
	}
	return snippets, nil
}
