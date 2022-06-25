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

	"github.com/spf13/afero"
)

var ErrTargetDirectoryMisnomer = errors.New("target directory is not named 'Deleted Games and Apps'")

// Organiser will organise media it finds under the given 'Deleted Games and Apps' directory.
type Organiser struct {
	fs       afero.Fs
	basePath string
	siblings []string
}

// New creates a new Organiser.
// If a target filesystem is not set with an option, the native OS will be used.
func New(options ...func(*Organiser) error) (*Organiser, error) {
	org := &Organiser{}

	for _, option := range options {
		if err := option(org); err != nil {
			return nil, err
		}
	}

	if org.basePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not get working directory: %w", err)
		}

		org.basePath = wd
	}

	if filepath.Base(org.basePath) != "Deleted Games and Apps" {
		return nil, ErrTargetDirectoryMisnomer
	}

	if org.fs == nil {
		org.fs = afero.NewOsFs()
	}

	prefixes, err := org.discover()
	if err != nil {
		return nil, err
	}

	org.siblings = prefixes

	return org, nil
}

// Path sets the base path for a new Organiser.
func Path(path string) func(*Organiser) error {
	return func(org *Organiser) error {
		org.basePath = path

		return nil
	}
}

// Filesystem overrides the filesystem that a new Organiser will operate on.
// If this option is not set, the Organiser will fall back to the native OS.
func Filesystem(fs afero.Fs) func(*Organiser) error {
	return func(org *Organiser) error {
		org.fs = fs

		return nil
	}
}

// GetNames returns a list of all app/game name prefixes discovered in the 'Deleted Games and Apps' base directory.
func (o *Organiser) GetNames() []string {
	duplicate := make([]string, len(o.siblings))
	copy(duplicate, o.siblings)

	return duplicate
}

// discover will search the Organiser's base path and return all of the app/game prefixes that it finds.
func (o *Organiser) discover() (prefixes []string, err error) {
	files, err := afero.ReadDir(o.fs, o.basePath)
	if err != nil {
		return nil, fmt.Errorf("could not read '%s': %w", o.basePath, err)
	}

	discovered := make([]string, 0)

	for _, file := range files {
		if file.Mode().IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		splitFilename := strings.SplitN(file.Name(), "_", 2)
		discovered = append(discovered, splitFilename[0])
	}

	return discovered, nil
}

// Prepare will create sibling directories for each discovered app/game name alongside the originating 'Deleted Games
// and Apps' base path.
func (o *Organiser) Prepare() error {
	for _, sibling := range o.siblings {
		create := filepath.Join(filepath.Dir(o.basePath), sibling)

		if err := o.fs.Mkdir(create, 0o777); err != nil {
			return fmt.Errorf("could not create directory '%s': %w", create, err)
		}
	}

	return nil
}

// Undelete moves the screenshots and video clips out of the 'Deleted Games and Apps' directory and into the
// app/game-specific directories that would have been created by calling o.Prepare.
func (o *Organiser) Undelete() error {
	sourceFiles, err := afero.ReadDir(o.fs, o.basePath)
	if err != nil {
		return fmt.Errorf("could not read directory '%s': %w", o.basePath, err)
	}

	for _, sibling := range o.siblings {
		destinationDir := filepath.Join(filepath.Dir(o.basePath), sibling)

		prefixMatches := make([]string, 0)

		for i := range sourceFiles {
			if strings.HasPrefix(sourceFiles[i].Name(), sibling) {
				prefixMatches = append(prefixMatches, sourceFiles[i].Name())
			}
		}

		for j := range prefixMatches {
			oldFilePath := filepath.Join(o.basePath, prefixMatches[j])
			newFilePath := filepath.Join(destinationDir, prefixMatches[j])

			if err := o.fs.Rename(oldFilePath, newFilePath); err != nil {
				return fmt.Errorf("could not move file '%s': %w", oldFilePath, err)
			}
		}
	}

	return nil
}
