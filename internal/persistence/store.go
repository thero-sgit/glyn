package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/thero-sgit/glyn/internal/config"
	_ "modernc.org/sqlite"
)

var migrations = []string{
	// v1
	`CREATE TABLE sessions (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		project_key TEXT NOT NULL,
		cwd 		TEXT NOT NULL,
		model 		TEXT NOT NULL,
		title       TEXT NOT NULL DEFAULT '',
		created_at  TEXT NOT NULL,
		updated_at  TEXT NOT NULL
	);
	CREATE TABLE messages (
		seq 		 INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id   INTEGER NOT NULL REFERENCES session(id) ON DELETE CASCADE,
		role 		 TEXT NOT NULL,
		content      TEXT,
		tool_calls   TEXT,
		tool_call_id TEXT,
		created_at   TEXT NOT NULL
	);
	CREATE INDEX idx_messages_session ON messages(session_id, seq);
	CREATE INDEX idx_sessions_project ON sessions(project_key, updated_at DESC);`,
}

type Store struct { db *sql.DB }

func Open(ctx context.Context) (*Store, error) {
	base := config.Config.DataBaseDir

	dsn := "file:" + filepath.Join(base, "glyn.db") +
		"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
 	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db,}, nil
}

func (s *Store) StartNewSession(ctx context.Context, title string) (ChatSession, error) {
	now := time.Now().Format("20060102-150405")

	result, err := s.db.ExecContext(ctx, 
		`INSERT INTO sessions (project_key, cwd, model, title, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		config.Config.ProjectKey, config.Config.WorkingDir, config.Config.Model.Name, title, now, now,
	)
	if err != nil {
		return ChatSession{}, err
	}

	id, _ := result.LastInsertId()
	return *NewChatSession(title, string(id), now), nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	var v int
	if err := db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&v); err != nil {
		return err
	}

	for i := v; i < len(migrations); i++ {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		
		if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migrations %d: %w", i+1, err)
		}

		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil	
}