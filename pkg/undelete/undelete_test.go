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
	_, err := undelete.New(undelete.OptionPath("/misnamed/directory"))

	// Assert
	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // unexpected error
}

func TestNewSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.New(undelete.OptionPath("testdata/populated/Deleted Games and Apps"))

	// Assert
	is.NoErr(err) // unexpected error
}

func TestNewSucceedsWhenCurrentDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	wd, err := os.Getwd()
	is.NoErr(err)
	t.Cleanup(func() { is.NoErr(os.Chdir(wd)) })
	is.NoErr(os.Chdir("testdata/populated/Deleted Games and Apps"))

	_, err = undelete.New()

	// Assert
	is.NoErr(err) // unexpected error
}

func TestDiscoverFindsZeroPrefixesWhenTargetDirectoryIsEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	o, err := undelete.New(undelete.OptionPath("testdata/unpopulated/Deleted Games and Apps"))
	is.NoErr(err)

	names, err := o.Discover()
	is.NoErr(err)

	// Assert
	is.Equal(len(names), 0)
}

func TestDiscoverFindsAtLeastOnePrefixWhenTargetDirectoryNotEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	org, err := undelete.New(undelete.OptionPath("testdata/populated/Deleted Games and Apps"))
	is.NoErr(err)
	names, err := org.Discover()
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
	org, err := undelete.New(undelete.OptionPath(tmpDir))
	is.NoErr(err)
	_, err = org.Discover()

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
	org, err := undelete.New(
		undelete.OptionFilesystem(testFS),
		undelete.OptionPath("testdata/populated/Deleted Games and Apps"),
	)
	is.NoErr(err)
	names, err := org.Discover()
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

func TestCreateErrorOnDirectoryPermissions(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Should switch from t.TempDir to afero.MemMapFs once these issues are resolved: //nolint:godox
	// https://github.com/spf13/afero/issues/150
	// https://github.com/spf13/afero/issues/335

	parentTmpDir := t.TempDir()
	tmpDir := filepath.Join(parentTmpDir, "Deleted Games and Apps")
	is.NoErr(os.Mkdir(tmpDir, 0o700))
	is.NoErr(os.WriteFile(filepath.Join(tmpDir, "prefix_123.png"), []byte("bytes"), 0o600))

	// Act
	org, err := undelete.New(undelete.OptionPath(tmpDir))
	is.NoErr(err)
	names, err := org.Discover()
	is.NoErr(err)
	is.Equal(len(names), 1)
	is.Equal(names[0], "prefix")

	// Remove permissions from parent directory, under which sibling(s) would be created
	is.NoErr(os.Chmod(parentTmpDir, 0o000))

	err = org.Create(names)

	// Assert
	is.True(errors.Is(err, fs.ErrPermission)) // should get a 'permission denied' error

	//// Prevent t.Cleanup from erroring when tidying up the directory that doesn't have any permissions
	is.NoErr(os.Chmod(parentTmpDir, 0o700)) //nolint:gosec // Need to clean up a directory, not a file
}
