package persistence

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	domain "github.com/rou-gan/WhiteKey-Experimentation-Site/backend/internal/domain/message"
)

type messageCSVRepository struct {
	path string
	mu   sync.RWMutex
}

func NewMessageCSVRepository(path string) domain.Repository {
	return &messageCSVRepository{path: path}
}

func (r *messageCSVRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Path is provided via trusted configuration (DATA_CSV_PATH), not request input.
	f, err := os.Open(r.path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("message not found: %s", id)
	}

	// Skip header row.
	for _, row := range records[1:] {
		if len(row) < 2 {
			continue
		}
		if domain.ID(row[0]) == id {
			return domain.NewMessage(domain.ID(row[0]), row[1])
		}
	}
	return nil, fmt.Errorf("message not found: %s", id)
}
