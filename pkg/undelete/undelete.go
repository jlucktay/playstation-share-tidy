// Package undelete will sort screenshots and video clips from the 'Deleted Games and Apps' directory into individual
// directories (that will be created as siblings of the 'Deleted ...' directory) based on all unique filename prefixes
// that it finds, as the prefix in front of the timestamp in the filename denotes the name of the game/app.
package undelete

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var ErrTargetDirectoryMisnomer = errors.New("target directory is not named 'Deleted Games and Apps'")

// Organiser will organise media it finds under the given 'Deleted Games and Apps' directory.
type Organiser struct {
	fs       afero.Fs
	basePath string
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

	if org.fs == nil {
		org.fs = afero.NewOsFs()
	}

	return org, nil
}

// OptionPath sets the base path for a new Organiser.
func OptionPath(path string) func(*Organiser) error {
	return func(org *Organiser) error {
		if filepath.Base(path) != "Deleted Games and Apps" {
			return ErrTargetDirectoryMisnomer
		}

		org.basePath = path

		return nil
	}
}

// OptionFilesystem overrides the filesystem that a new Organiser will operate on.
// If this option is not set, the Organiser will fall back to the native OS.
func OptionFilesystem(fs afero.Fs) func(*Organiser) error {
	return func(org *Organiser) error {
		org.fs = fs

		return nil
	}
}

// Discover will search the given target directory and return all of the app/game prefixes that it finds.
func (o *Organiser) Discover() (prefixes []string, err error) {
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

// Create the given sibling directories alongside the originating 'Deleted Games and Apps' directory.
func (o *Organiser) Create(siblings []string) error {
	for _, sibling := range siblings {
		create := filepath.Join(filepath.Dir(o.basePath), sibling)

		if err := o.fs.Mkdir(create, 0o777); err != nil {
			return fmt.Errorf("could not create directory '%s': %w", create, err)
		}
	}

	return nil
}
