package undelete_test

import (
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/undelete"
)

func TestUndeleteFailsWhenTargetDirectoryMisnamed(t *testing.T) {
	is := is.New(t)

	err := undelete.Run("/misnamed/directory")

	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // unexpected error
}

func TestUndeleteSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	is := is.New(t)

	err := undelete.Run("/a/b/c/Deleted Games and Apps")

	is.NoErr(err) // unexpected error
}
