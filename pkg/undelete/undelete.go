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

// New creates a new Organiser. If targetFS is given as 'nil' then the native OS will be used.
func New(targetFS afero.Fs, targetDir string) (*Organiser, error) {
	if targetFS == nil {
		targetFS = afero.NewOsFs()
	}

	if filepath.Base(targetDir) != "Deleted Games and Apps" {
		return nil, ErrTargetDirectoryMisnomer
	}

	return &Organiser{
		basePath: filepath.Clean(targetDir),
		fs:       targetFS,
	}, nil
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
