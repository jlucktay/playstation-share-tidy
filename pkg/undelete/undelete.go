// Package undelete will sort screenshots and video clips from the 'Deleted Games and Apps' directory into individual
// directories (that will be created as siblings of the 'Deleted ...' directory) based on all unique filename prefixes
// that it finds, as the prefix in front of the timestamp in the filename denotes the name of the game/app.
package undelete

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrTargetDirectoryMisnomer = errors.New("target directory is not named 'Deleted Games and Apps'")

// Organiser will organise media it finds under the given 'Deleted Games and Apps' directory.
type Organiser struct {
	basePath string
}

// New creates a new Organiser.
func New(targetDir string) (*Organiser, error) {
	if filepath.Base(targetDir) != "Deleted Games and Apps" {
		return nil, ErrTargetDirectoryMisnomer
	}

	return &Organiser{basePath: targetDir}, nil
}

// DiscoverPrefixes will search the given target directory and return all of the app/game prefixes that it finds.
func (o *Organiser) DiscoverPrefixes() ([]string, error) {
	files, err := os.ReadDir(o.basePath)
	if err != nil {
		return nil, fmt.Errorf("could not read '%s': %w", o.basePath, err)
	}

	discovered := make([]string, 0)

	for _, file := range files {
		if file.Type().IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		splitFilename := strings.SplitN(file.Name(), "_", 2)
		discovered = append(discovered, splitFilename[0])
	}

	return discovered, nil
}

// Create the given sibling directories alongside the originating 'Deleted Games and Apps' directory.
func (o *Organiser) Create(siblings []string) error {
	return nil
}
