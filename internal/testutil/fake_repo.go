package testutil

import (
	"context"
	"strings"

	"github.com/elfoundation/hatch/internal/store"
)

// FakeRepository implements store.Repository for tests.
type FakeRepository struct {
	Endpoints map[string]*store.Endpoint
	Requests  []*store.Request
	Mocks     map[string]*store.MockConfig
}

// NewFakeRepository creates a new FakeRepository.
func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Endpoints: map[string]*store.Endpoint{},
		Mocks:     map[string]*store.MockConfig{},
	}
}

func (f *FakeRepository) CreateEndpoint(_ context.Context, url string) (*store.Endpoint, error) {
	e := &store.Endpoint{ID: url, URL: url, CreatedAt: "t", UpdatedAt: "t"}
	f.Endpoints[url] = e
	return e, nil
}

func (f *FakeRepository) GetEndpoint(_ context.Context, id string) (*store.Endpoint, error) {
	e, ok := f.Endpoints[id]
	if !ok {
		return nil, errNotFound
	}
	return e, nil
}

func (f *FakeRepository) AppendRequest(_ context.Context, eid string, r *store.Request) error {
	r.ID = "req-" + string(rune(len(f.Requests)+'0'))
	r.EndpointID = eid
	f.Requests = append(f.Requests, r)
	return nil
}

func (f *FakeRepository) GetRequest(_ context.Context, id string) (*store.Request, error) {
	for _, r := range f.Requests {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errNotFound
}

func (f *FakeRepository) ListRequests(_ context.Context, _ string, _ int) ([]*store.Request, error) {
	return f.Requests, nil
}

func (f *FakeRepository) SearchRequests(_ context.Context, _ string, query string, limit int) ([]*store.Request, error) {
	if query == "" {
		return f.Requests, nil
	}
	var matched []*store.Request
	for _, r := range f.Requests {
		if strings.Contains(r.Method, query) ||
			strings.Contains(r.Path, query) ||
			strings.Contains(r.Headers, query) ||
			strings.Contains(r.Query, query) ||
			strings.Contains(string(r.Body), query) {
			matched = append(matched, r)
			if limit > 0 && len(matched) >= limit {
				break
			}
		}
	}
	return matched, nil
}

func (f *FakeRepository) GetMock(_ context.Context, endpointID string) (*store.MockConfig, error) {
	m, ok := f.Mocks[endpointID]
	if !ok {
		return nil, errNotFound
	}
	return m, nil
}

func (f *FakeRepository) SetMock(_ context.Context, mock *store.MockConfig) error {
	f.Mocks[mock.EndpointID] = mock
	return nil
}

func (f *FakeRepository) Close() error { return nil }

func (f *FakeRepository) Ping(_ context.Context) error { return nil }

type notFoundError struct{}

func (e *notFoundError) Error() string { return "not found" }

var errNotFound = &notFoundError{}
