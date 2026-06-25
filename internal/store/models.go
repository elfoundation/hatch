package store

import "context"

type Endpoint struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Request struct {
	ID         string `json:"id"`
	EndpointID string `json:"endpoint_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Headers    string `json:"headers"`
	Query      string `json:"query"`
	Body       []byte `json:"body"`
	CreatedAt  string `json:"created_at"`
}

// MockConfig holds the mock response configuration for an endpoint.
type MockConfig struct {
	EndpointID string            `json:"endpoint_id"`
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

type Repository interface {
	CreateEndpoint(ctx context.Context, url string) (*Endpoint, error)
	GetEndpoint(ctx context.Context, id string) (*Endpoint, error)
	AppendRequest(ctx context.Context, endpointID string, req *Request) error
	GetRequest(ctx context.Context, id string) (*Request, error)
	ListRequests(ctx context.Context, endpointID string, limit int) ([]*Request, error)
	SearchRequests(ctx context.Context, endpointID string, query string, limit int) ([]*Request, error)
	GetMock(ctx context.Context, endpointID string) (*MockConfig, error)
	SetMock(ctx context.Context, mock *MockConfig) error
	Close() error
	Ping(ctx context.Context) error
}

var _ Repository = (*sqliteRepo)(nil)
