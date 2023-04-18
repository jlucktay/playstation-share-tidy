package undelete_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/undelete"
)

func TestNewErrorsOnMissingWorkingDirectory(t *testing.T) {
	// Arrange
	is := is.New(t)

	wd, err := os.Getwd()
	is.NoErr(err) // can't get working directory

	tmp, err := os.MkdirTemp(filepath.Join(wd, "testdata"), t.Name()+"-temporary")
	is.NoErr(err) // can't create temp directory

	t.Cleanup(func() {
		is.NoErr(os.RemoveAll(tmp)) // cleanup can't remove temp directory
	})

	basePath := filepath.Join(tmp, "Deleted Games and Apps")

	is.NoErr(os.MkdirAll(basePath, 0o750)) // can't create temp base path
	is.NoErr(os.Chdir(basePath))           // can't change working directory to temp base path

	t.Cleanup(func() {
		is.NoErr(os.Chdir(wd)) // cleanup can't revert working directory
	})

	is.NoErr(os.RemoveAll(tmp)) // can't remove temp directory

	// Act
	_, err = undelete.New()

	// Assert
	is.True(err != nil) // undelete.New should error when working directory does not exist

	var target *fs.PathError

	is.True(errors.As(err, &target))     // error should be of type *fs.PathError
	is.Equal(target.Err, syscall.ENOENT) // no such file or directory
	is.Equal(target.Op, "open")          // operation should be 'open'
	is.Equal(target.Path, basePath)      // path should match temp base path
}
