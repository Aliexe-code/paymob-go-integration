package payment

import (
	"os"
)

// NewTestRepository creates an in-memory repository for testing
// This function is exported for use in other package tests
func NewTestRepository() (*Repository, func()) {
	// Create temp file for SQLite
	f, err := os.CreateTemp("", "test-payments-*.db")
	if err != nil {
		panic(err)
	}
	path := f.Name()
	f.Close()

	repo, err := NewRepository(path)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		repo.Close()
		os.Remove(path)
	}

	return repo, cleanup
}
