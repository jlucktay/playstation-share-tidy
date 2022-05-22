package dedupe_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/dedupe"
)

func TestLifecycle(t *testing.T) {
	is := is.New(t)

	d, err := dedupe.New(".")
	is.NoErr(err)
	is.NoErr(d.Close())
}

func TestNewCannotCallNewWithNonExistentDirectory(t *testing.T) {
	is := is.New(t)
	_, err := dedupe.New("non-existent directory")
	is.True(errors.Is(err, fs.ErrNotExist))
}
