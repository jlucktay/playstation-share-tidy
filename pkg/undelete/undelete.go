// Package undelete will sort screenshots and video clips from the 'Deleted Games and Apps' directory into individual
// directories (that will be created as siblings of the 'Deleted ...' directory) based on all unique filename prefixes
// that it finds, as the prefix in front of the timestamp in the filename denotes the name of the game/app.
package undelete

import (
	"errors"
	"path/filepath"
)

var ErrTargetDirectoryMisnomer = errors.New("target directory is not named 'Deleted Games and Apps'")

// Run the undelete process on the given target directory.
func Run(targetDir string) error {
	if filepath.Base(targetDir) != "Deleted Games and Apps" {
		return ErrTargetDirectoryMisnomer
	}

	return nil
}
