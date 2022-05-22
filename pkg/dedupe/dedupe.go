//go:generate mockgen -source=dedupe.go -destination=./mock_dedupe_test.go -package=dedupe_test

// Package dedupe provides a method to deduplicate files in a given directory.
package dedupe

import (
	"fmt"
	"os"
)

// Deduplicator will walk the directory it is opened on, checksumming files with similar names, and marking the
// duplicate files.
type Deduplicator struct{}

// New creates a new uninitialised Deduplicator for the given directory, in preparation for walking it.
func New(directory string) (*Deduplicator, error) {
	if _, err := os.Stat(directory); err != nil {
		return nil, fmt.Errorf("could not stat directory '%s': %w", directory, err)
	}

	return &Deduplicator{}, nil
}

// Close the Deduplicator once finished with it.
func (d *Deduplicator) Close() error {
	return nil
}
