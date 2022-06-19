package undelete_test

import (
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/undelete"
)

func TestDiscoverFailsWhenTargetDirectoryMisnamed(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.DiscoverPrefixes("/misnamed/directory")

	// Assert
	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // unexpected error
}

func TestDiscoverSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.DiscoverPrefixes("testdata/populated/Deleted Games and Apps")

	// Assert
	is.NoErr(err) // unexpected error
}

func TestDiscoverFindsZeroPrefixesWhenTargetDirectoryIsEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	names, err := undelete.DiscoverPrefixes("testdata/unpopulated/Deleted Games and Apps")

	// Assert
	is.NoErr(err)
	is.Equal(len(names), 0)
}

func TestDiscoverFindsAtLeastOnePrefixWhenTargetDirectoryNotEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	names, err := undelete.DiscoverPrefixes("testdata/populated/Deleted Games and Apps")

	// Assert
	is.NoErr(err)
	is.Equal(len(names), 3)
	is.Equal(names[0], "Bugsnax")
	is.Equal(names[1], "Control Ultimate Edition")
	is.Equal(names[2], "DEATH STRANDING DIRECTOR'S CUT")
}
