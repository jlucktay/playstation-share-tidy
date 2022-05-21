//go:generate mockgen -source=dedupe.go -destination=./mock_dedupe_test.go -package=dedupe_test

// Package dedupe provides a method to deduplicate files in a given directory.
package dedupe

// Deduplicator will walk the directory it is opened on, checksumming files with similar names, and marking the
// duplicate files.
type Deduplicator struct{}

// New creates a new uninitialised Deduplicator for the given directory, in preparation for walking it.
func New(directory string) (*Deduplicator, error) {
	return &Deduplicator{}, nil
}

// Close the Deduplicator once finished with it.
func (d *Deduplicator) Close() error {
	return nil
}
