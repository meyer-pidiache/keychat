package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opc/keychat/internal/relay"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id         TEXT PRIMARY KEY,
		pubkey     TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		kind       INTEGER NOT NULL,
		content    TEXT NOT NULL,
		sig        TEXT NOT NULL,
		stored_at  TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS event_tags (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
		key      TEXT NOT NULL,
		value    TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_events_pubkey ON events(pubkey);
	CREATE INDEX IF NOT EXISTS idx_events_kind ON events(kind);
	CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);
	CREATE INDEX IF NOT EXISTS idx_tags_key_value ON event_tags(key, value);
	CREATE INDEX IF NOT EXISTS idx_tags_event_id ON event_tags(event_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) SaveEvent(event *relay.Event) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT OR IGNORE INTO events (id, pubkey, created_at, kind, content, sig) VALUES (?, ?, ?, ?, ?, ?)`,
		event.ID, event.Pubkey, event.CreatedAt, event.Kind, event.Content, event.Sig)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO event_tags (event_id, key, value) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			stmt.Exec(event.ID, tag[0], tag[1])
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) QueryEvents(filters []relay.Filter) ([]*relay.Event, error) {
	if len(filters) == 0 {
		return nil, nil
	}
	seen := make(map[string]*relay.Event)
	for _, f := range filters {
		events, err := s.queryFilter(&f)
		if err != nil {
			return nil, err
		}
		for _, e := range events {
			seen[e.ID] = e
		}
	}
	result := make([]*relay.Event, 0, len(seen))
	for _, e := range seen {
		result = append(result, e)
	}
	return result, nil
}

func (s *SQLiteStore) queryFilter(f *relay.Filter) ([]*relay.Event, error) {
	query := "SELECT DISTINCT e.id, e.pubkey, e.created_at, e.kind, e.content, e.sig FROM events e"
	var args []interface{}
	var conditions []string
	argIdx := 1

	if len(f.IDs) > 0 {
		placeholders := make([]string, len(f.IDs))
		for i, id := range f.IDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("e.id IN (%s)", joinPlaceholders(placeholders)))
	}
	if len(f.Authors) > 0 {
		placeholders := make([]string, len(f.Authors))
		for i, pubkey := range f.Authors {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, pubkey)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("e.pubkey IN (%s)", joinPlaceholders(placeholders)))
	}
	if len(f.Kinds) > 0 {
		placeholders := make([]string, len(f.Kinds))
		for i, kind := range f.Kinds {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, kind)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("e.kind IN (%s)", joinPlaceholders(placeholders)))
	}
	for key, vals := range f.Tags {
		placeholders := make([]string, len(vals))
		keyIdx := argIdx
		args = append(args, key)
		argIdx++
		for i, v := range vals {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, v)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf(
			`e.id IN (SELECT event_id FROM event_tags WHERE key = $%d AND value IN (%s))`,
			keyIdx, joinPlaceholders(placeholders)))
	}
	if f.Since != nil {
		conditions = append(conditions, fmt.Sprintf("e.created_at >= $%d", argIdx))
		args = append(args, *f.Since)
		argIdx++
	}
	if f.Until != nil {
		conditions = append(conditions, fmt.Sprintf("e.created_at <= $%d", argIdx))
		args = append(args, *f.Until)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions, " AND ")
	}

	query += " ORDER BY e.created_at DESC"

	if f.Limit != nil && *f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, *f.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var events []*relay.Event
	for rows.Next() {
		var e relay.Event
		if err := rows.Scan(&e.ID, &e.Pubkey, &e.CreatedAt, &e.Kind, &e.Content, &e.Sig); err != nil {
			return nil, err
		}
		tags, err := s.getTags(e.ID)
		if err != nil {
			return nil, err
		}
		e.Tags = tags
		events = append(events, &e)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) getTags(eventID string) ([]relay.Tag, error) {
	rows, err := s.db.Query("SELECT key, value FROM event_tags WHERE event_id = $1 ORDER BY id", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagMap := make(map[string]relay.Tag)
	var order []string
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		tagMap[key] = append(tagMap[key], key, value)
		order = append(order, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var tags []relay.Tag
	seen := make(map[string]bool)
	for _, k := range order {
		if !seen[k] {
			seen[k] = true
			tags = append(tags, tagMap[k])
		}
	}
	return tags, nil
}

func (s *SQLiteStore) DeleteOlderThan(t time.Time) (int64, error) {
	result, err := s.db.Exec("DELETE FROM events WHERE created_at < ?", t.Unix())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func joinPlaceholders(ps []string) string {
	result := ps[0]
	for _, p := range ps[1:] {
		result += ", " + p
	}
	return result
}

func joinConditions(conds []string, sep string) string {
	result := conds[0]
	for _, c := range conds[1:] {
		result += sep + " " + c
	}
	return result
}

func tagsToJSON(tags []relay.Tag) string {
	data, _ := json.Marshal(tags)
	return string(data)
}
