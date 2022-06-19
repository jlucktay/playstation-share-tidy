package undelete_test

import (
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/undelete"
)

func TestUndeleteFailsWhenTargetDirectoryMisnamed(t *testing.T) {
	is := is.New(t)

	_, err := undelete.DiscoverPrefixes("/misnamed/directory")

	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // unexpected error
}

func TestUndeleteSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	is := is.New(t)

	_, err := undelete.DiscoverPrefixes("/a/b/c/Deleted Games and Apps")

	is.NoErr(err) // unexpected error
}

func TestUndeleteFindsAtLeastOnePrefixWhenTargetDirectoryNotEmpty(t *testing.T) {
	is := is.New(t)

	names, err := undelete.DiscoverPrefixes("testdata/populated/Deleted Games and Apps")
	is.NoErr(err)
	is.True(len(names) > 0)
}
