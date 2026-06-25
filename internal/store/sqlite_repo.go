package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type sqliteRepo struct{ db *sql.DB }

func NewSQLiteRepo(db *sql.DB) (Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("store: nil db")
	}
	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) CreateEndpoint(ctx context.Context, url string) (*Endpoint, error) {
	now := utcNow()
	e := &Endpoint{ID: url, URL: url, CreatedAt: now, UpdatedAt: now}
	_, err := r.db.ExecContext(ctx, `INSERT INTO endpoints (id, url, created_at, updated_at) VALUES (?, ?, ?, ?)`, e.ID, e.URL, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("store: create endpoint: %w", err)
	}
	return e, nil
}

func (r *sqliteRepo) GetEndpoint(ctx context.Context, id string) (*Endpoint, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, url, created_at, updated_at FROM endpoints WHERE id = ?`, id)
	e := &Endpoint{}
	if err := row.Scan(&e.ID, &e.URL, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, fmt.Errorf("store: get endpoint: %w", err)
	}
	return e, nil
}

func (r *sqliteRepo) AppendRequest(ctx context.Context, endpointID string, req *Request) error {
	req.ID = uuid.NewString()
	req.EndpointID = endpointID
	req.CreatedAt = utcNow()
	_, err := r.db.ExecContext(ctx, `INSERT INTO requests (id, endpoint_id, method, path, headers, query, body, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ID, req.EndpointID, req.Method, req.Path, req.Headers, req.Query, req.Body, req.CreatedAt)
	if err != nil {
		return fmt.Errorf("store: append request: %w", err)
	}
	return nil
}

func (r *sqliteRepo) GetRequest(ctx context.Context, id string) (*Request, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, endpoint_id, method, path, headers, query, body, created_at FROM requests WHERE id = ?`, id)
	req := &Request{}
	if err := row.Scan(&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers, &req.Query, &req.Body, &req.CreatedAt); err != nil {
		return nil, fmt.Errorf("store: get request %s: %w", id, err)
	}
	return req, nil
}

func (r *sqliteRepo) ListRequests(ctx context.Context, endpointID string, limit int) ([]*Request, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, endpoint_id, method, path, headers, query, body, created_at FROM requests WHERE endpoint_id = ? ORDER BY created_at DESC LIMIT ?`, endpointID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list requests: %w", err)
	}
	defer rows.Close()
	var out []*Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers, &req.Query, &req.Body, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan request: %w", err)
		}
		out = append(out, &req)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) GetMock(ctx context.Context, endpointID string) (*MockConfig, error) {
	row := r.db.QueryRowContext(ctx, `SELECT mock_status, mock_headers, mock_body FROM endpoints WHERE id = ?`, endpointID)
	var status sql.NullInt64
	var headersJSON sql.NullString
	var body []byte
	if err := row.Scan(&status, &headersJSON, &body); err != nil {
		return nil, fmt.Errorf("store: get mock: %w", err)
	}
	if !status.Valid {
		return nil, fmt.Errorf("store: no mock configured for %s", endpointID)
	}
	m := &MockConfig{
		EndpointID: endpointID,
		Status:     int(status.Int64),
		Body:       body,
	}
	if headersJSON.Valid && headersJSON.String != "" {
		json.Unmarshal([]byte(headersJSON.String), &m.Headers)
	}
	return m, nil
}

func (r *sqliteRepo) SetMock(ctx context.Context, mock *MockConfig) error {
	headersJSON, err := json.Marshal(mock.Headers)
	if err != nil {
		return fmt.Errorf("store: marshal mock headers: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE endpoints SET mock_status = ?, mock_headers = ?, mock_body = ? WHERE id = ?`,
		mock.Status, string(headersJSON), mock.Body, mock.EndpointID,
	)
	if err != nil {
		return fmt.Errorf("store: set mock: %w", err)
	}
	return nil
}

func (r *sqliteRepo) SearchRequests(ctx context.Context, endpointID string, query string, limit int) ([]*Request, error) {
	if limit <= 0 {
		limit = 50
	}
	// Simple LIKE search across text fields. Convert body to text.
	rows, err := r.db.QueryContext(ctx, `SELECT id, endpoint_id, method, path, headers, query, body, created_at FROM requests WHERE endpoint_id = ? AND (method LIKE ? OR path LIKE ? OR headers LIKE ? OR query LIKE ? OR CAST(body AS TEXT) LIKE ?) ORDER BY created_at DESC LIMIT ?`, endpointID, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("store: search requests: %w", err)
	}
	defer rows.Close()
	var out []*Request
	for rows.Next() {
		var req Request
		if err := rows.Scan(&req.ID, &req.EndpointID, &req.Method, &req.Path, &req.Headers, &req.Query, &req.Body, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan request: %w", err)
		}
		out = append(out, &req)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) Close() error { return r.db.Close() }

func (r *sqliteRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func utcNow() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00") }
