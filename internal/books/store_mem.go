package books

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound          = errors.New("book not found")
	ErrInvalidName       = errors.New("name is required")
	ErrReadPageTooBig    = errors.New("readPage must be <= pageCount")
)

type Store interface {
	Create(b *Book) (string, error)
	Get(id string) (*Book, error)
	List(filter Filter) ([]Book, error)
	Update(id string, patch Book) error
	Delete(id string) error
}

type Filter struct {
	Name     string
	Reading  *bool // nil = ignore
	Finished *bool // nil = ignore
}

type memStore struct {
	mu    sync.RWMutex
	items map[string]Book
	idSeq int64
}

func NewMemStore() Store {
	return &memStore{items: make(map[string]Book)}
}

func (m *memStore) nextID() string {
	m.idSeq++
	return randomID() // stable random id helper below
}

func (m *memStore) Create(b *Book) (string, error) {
	if strings.TrimSpace(b.Name) == "" {
		return "", ErrInvalidName
	}
	if b.ReadPage > b.PageCount {
		return "", ErrReadPageTooBig
	}

	now := time.Now()
	b.ID = m.nextID()
	b.Finished = b.PageCount == b.ReadPage
	b.InsertedAt = now
	b.UpdatedAt = now

	m.mu.Lock()
	m.items[b.ID] = *b
	m.mu.Unlock()
	return b.ID, nil
}

func (m *memStore) Get(id string) (*Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &v, nil
}

func (m *memStore) List(f Filter) ([]Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Book, 0, len(m.items))
	needle := strings.ToLower(strings.TrimSpace(f.Name))

	for _, b := range m.items {
		if needle != "" && !strings.Contains(strings.ToLower(b.Name), needle) {
			continue
		}
		if f.Reading != nil && b.Reading != *f.Reading {
			continue
		}
		if f.Finished != nil && b.Finished != *f.Finished {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (m *memStore) Update(id string, patch Book) error {
	if strings.TrimSpace(patch.Name) == "" {
		return ErrInvalidName
	}
	if patch.ReadPage > patch.PageCount {
		return ErrReadPageTooBig
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.items[id]
	if !ok {
		return ErrNotFound
	}
	old.Name = patch.Name
	old.Author = patch.Author
	old.Publisher = patch.Publisher
	old.PageCount = patch.PageCount
	old.ReadPage = patch.ReadPage
	old.Reading = patch.Reading
	old.Finished = patch.PageCount == patch.ReadPage
	old.UpdatedAt = time.Now()
	m.items[id] = old
	return nil
}

func (m *memStore) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

// --- helpers ---

// simple URL-safe random id
func randomID() string {
	const alnum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	b := make([]byte, 16)
	for i := range b {
		b[i] = alnum[randInt(len(alnum))]
	}
	return string(b)
}

// cheap PRNG ok for demo
var seed uint64 = 88172645463393265

func randInt(n int) int {
	// xorshift*
	seed ^= seed << 7
	seed ^= seed >> 9
	return int(seed % uint64(n))
}
