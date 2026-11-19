package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/thero-sgit/glyn/internal/ai/common"
	co "github.com/thero-sgit/glyn/internal/common"
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
		session_id   INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		role 		 TEXT NOT NULL,
		content      TEXT,
		tool_calls   BOOLEAN,
		tool_call_id TEXT,
		created_at   TEXT NOT NULL
	);
	CREATE TABLE tool_calls (
		id          TEXT PRIMARY KEY NOT NULL,
		idx         INTEGER NOT NULL,
		type        TEXT NOT NULL,
		name        TEXT NOT NULL,
		args        TEXT NOT NULL,
		message_seq INTEGER NOT NULL REFERENCES messages(seq) ON DELETE CASCADE
	);
	CREATE INDEX idx_tool_calls_message ON 	tool_calls(message_seq, id);
	CREATE INDEX idx_messages_session ON messages(session_id, seq);
	CREATE INDEX idx_sessions_project ON sessions(project_key, updated_at DESC);`,
}

type Store struct{ db *sql.DB }

func (s *Store) Close() error {
	return s.db.Close()
}

func Open(ctx context.Context) (*Store, error) {
	base := config.Config.DataBaseDir

	path := filepath.ToSlash(filepath.Join(base, "glyn.db"))
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)",
	}).String()
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

	return &Store{db: db}, nil
}

func (s *Store) ListSessions(ctx context.Context) ([]ChatSession, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, model, title, updated_at
		 FROM sessions
		 WHERE project_key = ?
		 ORDER BY updated_at DESC`,
		config.Config.ProjectKey,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.Id, &s.Model, &s.Name, &s.LastUpdated); err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, rows.Err()
}

func (s *Store) StartNewSession(ctx context.Context, title string) (ChatSession, error) {
	now := time.Now().Format(co.TsFormat)

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sessions (project_key, cwd, model, title, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?);`,
		config.Config.ProjectKey, config.Config.WorkingDir, config.Config.Model.Name, title, now, now,
	)
	if err != nil {
		return ChatSession{}, err
	}

	id, _ := result.LastInsertId()
	return *NewChatSession(title, fmt.Sprint(id), config.Config.Model.Name, now), nil
}

func (s *Store) AppendMessage(ctx context.Context, sessionId string, wrapper common.MessageWrapper) error {
	hasToolCalls := len(wrapper.Msg.ToolCalls) > 0

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO messages (session_id, role, content, tool_calls, tool_call_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?);`,
		sessionId, wrapper.Msg.Role, wrapper.Msg.Content, hasToolCalls, wrapper.Msg.ToolCallId, wrapper.Time,
	)

	if err != nil {
		return err
	}

	if hasToolCalls {
		seq, _ := result.LastInsertId()
		return s.appendToolCalls(ctx, seq, wrapper.Msg.ToolCalls)
	}

	return nil
}

func (s *Store) ListMessagesInSession(ctx context.Context, sessionId string) ([]common.MessageWrapper, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT seq, role, content, tool_calls, tool_call_id, created_at
		 FROM messages
		 WHERE session_id = ?
		 ORDER BY created_at ASC, seq ASC;`,
		sessionId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []common.MessageWrapper
	var messageSeqs []int64
	var hasToolCalls []bool
	for rows.Next() {
		var wrapper common.MessageWrapper
		var messageSeq int64
		var hasToolCallsFlag int8

		if err := rows.Scan(
			&messageSeq,
			&wrapper.Msg.Role,
			&wrapper.Msg.Content,
			&hasToolCallsFlag,
			&wrapper.Msg.ToolCallId,
			&wrapper.Time,
		); err != nil {
			return nil, err
		}

		out = append(out, wrapper)
		messageSeqs = append(messageSeqs, messageSeq)
		hasToolCalls = append(hasToolCalls, hasToolCallsFlag != 0)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for i, hasCalls := range hasToolCalls {
		if !hasCalls {
			continue
		}
		out[i].Msg.ToolCalls, err = s.listToolCalls(ctx, messageSeqs[i])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (s *Store) listToolCalls(ctx context.Context, message_seq int64) ([]common.ToolCall, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, idx, type, name, args
		 FROM tool_calls
		 WHERE message_seq = ?
		 ORDER BY idx ASC, id ASC`,
		message_seq,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []common.ToolCall
	for rows.Next() {
		var call common.ToolCall
		if err := rows.Scan(&call.Id, &call.Index, &call.Type, &call.Function.Name, &call.Function.Arguments); err != nil {
			return nil, err
		}

		out = append(out, call)
	}

	return out, rows.Err()
}

func (s *Store) appendToolCalls(ctx context.Context, seq int64, calls []common.ToolCall) error {
	for _, call := range calls {
		if _, err := s.db.ExecContext(
			ctx,
			`INSERT INTO tool_calls (id, idx, type, name, args, message_seq)
			VALUES (?, ?, ?, ?, ?, ?);`,
			call.Id, call.Index, call.Type, call.Function.Name, call.Function.Arguments, seq,
		); err != nil {
			return err
		}
	}

	return nil
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
