package undelete_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/matryer/is"
	"github.com/spf13/afero"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/undelete"
)

func TestNewFailsWhenTargetDirectoryMisnamed(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.New(nil, "/misnamed/directory")

	// Assert
	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // unexpected error
}

func TestNewSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.New(nil, "testdata/populated/Deleted Games and Apps")

	// Assert
	is.NoErr(err) // unexpected error
}

func TestDiscoverFindsZeroPrefixesWhenTargetDirectoryIsEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	o, err := undelete.New(nil, "testdata/unpopulated/Deleted Games and Apps")
	is.NoErr(err)

	names, err := o.DiscoverPrefixes()
	is.NoErr(err)

	// Assert
	is.Equal(len(names), 0)
}

func TestDiscoverFindsAtLeastOnePrefixWhenTargetDirectoryNotEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	org, err := undelete.New(nil, "testdata/populated/Deleted Games and Apps")
	is.NoErr(err)
	names, err := org.DiscoverPrefixes()
	is.NoErr(err)

	// Assert
	is.Equal(len(names), 3)
	is.Equal(names[0], "Bugsnax")
	is.Equal(names[1], "Control Ultimate Edition")
	is.Equal(names[2], "DEATH STRANDING DIRECTOR'S CUT")
}

func TestDiscoverErrorOnUnreadableDirectory(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Should switch from t.TempDir to afero.MemMapFs once these issues are resolved: //nolint:godox
	// https://github.com/spf13/afero/issues/150
	// https://github.com/spf13/afero/issues/335

	tmpDir := filepath.Join(t.TempDir(), "Deleted Games and Apps")
	is.NoErr(os.Mkdir(tmpDir, 0o700))
	is.NoErr(os.WriteFile(filepath.Join(tmpDir, "unreachable_123.png"), []byte("unreachable"), 0o600))
	is.NoErr(os.Chmod(tmpDir, 0o000))

	// Act
	org, err := undelete.New(nil, tmpDir)
	is.NoErr(err)
	_, err = org.DiscoverPrefixes()

	// Assert
	is.True(errors.Is(err, fs.ErrPermission)) // should get a 'permission denied' error

	//// Prevent t.Cleanup from erroring when tidying up the directory that doesn't have any permissions
	is.NoErr(os.Chmod(tmpDir, 0o700)) //nolint:gosec // Need to clean up a directory, not a file
}

func TestCreateWillMakeOneDirectoryPerSibling(t *testing.T) {
	// Arrange
	is := is.New(t)

	baseFS := afero.NewOsFs()
	readonlyBase := afero.NewReadOnlyFs(baseFS)
	testFS := afero.NewCopyOnWriteFs(readonlyBase, afero.NewMemMapFs())

	// Act
	org, err := undelete.New(testFS, "testdata/populated/Deleted Games and Apps")
	is.NoErr(err)
	names, err := org.DiscoverPrefixes()
	is.NoErr(err)
	err = org.Create(names)
	is.NoErr(err)

	// Assert
	for _, prefix := range []string{"Bugsnax", "Control Ultimate Edition", "DEATH STRANDING DIRECTOR'S CUT"} {
		dir, err := testFS.Stat(filepath.Join("testdata/populated", prefix))
		is.NoErr(err)        // could not stat new sibling directory
		is.True(dir.IsDir()) // sibling is not a directory
	}
}
