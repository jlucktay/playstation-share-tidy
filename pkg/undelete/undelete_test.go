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
	_, err := undelete.New(undelete.Path("/misnamed/directory"))

	// Assert
	is.Equal(err, undelete.ErrTargetDirectoryMisnomer) // error different from expected
}

func TestNewSucceedsWhenTargetDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	_, err := undelete.New(undelete.Path("testdata/populated/Deleted Games and Apps"))

	// Assert
	is.NoErr(err) // unexpected error
}

func TestNewSucceedsWhenCurrentDirectoryNamedCorrectly(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	wd, err := os.Getwd()
	is.NoErr(err) // can't get working directory
	t.Cleanup(func() {
		is.NoErr(os.Chdir(wd)) // cleanup can't revert working directory
	})
	is.NoErr(os.Chdir("testdata/populated/Deleted Games and Apps")) // can't change working directory

	_, err = undelete.New()

	// Assert
	is.NoErr(err) // unexpected error
}

func TestDiscoverFindsZeroPrefixesWhenTargetDirectoryIsEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	org, err := undelete.New(undelete.Path("testdata/unpopulated/Deleted Games and Apps"))
	is.NoErr(err) // could not create new Organiser

	// Assert
	is.Equal(len(org.GetNames()), 0) // unpopulated testdata directory should be empty
}

func TestDiscoverFindsAtLeastOnePrefixWhenTargetDirectoryNotEmpty(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Act
	org, err := undelete.New(undelete.Path("testdata/populated/Deleted Games and Apps"))
	is.NoErr(err) // could not create new Organiser

	names := org.GetNames()

	// Assert
	is.Equal(len(names), 3) // populated testdata should have three files
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
	is.NoErr(os.Mkdir(tmpDir, 0o700)) // could not create new directory under temp
	is.NoErr(os.WriteFile(filepath.Join(tmpDir, "unreachable_123.png"),
		[]byte("unreachable"), 0o600)) // could not write file into temp directory
	is.NoErr(os.Chmod(tmpDir, 0o000)) // could not change permissions on temp directory
	t.Cleanup(func() {
		//// Prevent t.Cleanup from erroring when tidying up the directory that doesn't have any permissions
		is.NoErr(
			os.Chmod(tmpDir, 0o700), //nolint:gosec // Need to clean up a directory, not a file
		) // could not revert permissions on temp directory
	})

	// Act
	_, err := undelete.New(undelete.Path(tmpDir))

	// Assert
	is.True(errors.Is(err, fs.ErrPermission)) // should get a 'permission denied' error
}

func TestPrepareWillMakeOneDirectoryPerSibling(t *testing.T) {
	// Arrange
	is := is.New(t)

	baseFS := afero.NewOsFs()
	readonlyBase := afero.NewReadOnlyFs(baseFS)
	testFS := afero.NewCopyOnWriteFs(readonlyBase, afero.NewMemMapFs())

	// Act
	org, err := undelete.New(
		undelete.Filesystem(testFS),
		undelete.Path("testdata/populated/Deleted Games and Apps"),
	)
	is.NoErr(err) // could not create new Organiser
	err = org.Prepare()
	is.NoErr(err) // error while creating sibling directories

	// Assert
	for _, prefix := range []string{"Bugsnax", "Control Ultimate Edition", "DEATH STRANDING DIRECTOR'S CUT"} {
		dir, err := testFS.Stat(filepath.Join("testdata/populated", prefix))
		is.NoErr(err)        // could not stat new sibling directory
		is.True(dir.IsDir()) // sibling is not a directory
	}
}

func TestPrepareErrorOnDirectoryPermissions(t *testing.T) {
	// Arrange
	is := is.New(t)

	// Should switch from t.TempDir to afero.MemMapFs once these issues are resolved:
	// https://github.com/spf13/afero/issues/150
	// https://github.com/spf13/afero/issues/335

	parentTmpDir := t.TempDir()
	tmpDir := filepath.Join(parentTmpDir, "Deleted Games and Apps")
	is.NoErr(os.Mkdir(tmpDir, 0o700)) // could not create temp directory
	is.NoErr(os.WriteFile(filepath.Join(tmpDir, "prefix_123.png"),
		[]byte("bytes"), 0o600)) // problem writing file to temp directory

	// Act
	org, err := undelete.New(undelete.Path(tmpDir))
	is.NoErr(err) // could not create new Organiser

	names := org.GetNames()
	is.Equal(len(names), 1) // should have discovered one prefix only
	is.Equal(names[0], "prefix")

	// Remove permissions from parent directory, under which sibling(s) would be created
	is.NoErr(os.Chmod(parentTmpDir, 0o000)) // could not set permissions on parent temp directory
	t.Cleanup(func() {
		//// Prevent t.Cleanup from erroring when tidying up the directory that doesn't have any permissions
		is.NoErr(
			os.Chmod(parentTmpDir, 0o700), //nolint:gosec // Need to clean up a directory, not a file
		) // could not revert permissions on temp directory
	})

	err = org.Prepare()

	// Assert
	is.True(errors.Is(err, fs.ErrPermission)) // should get a 'permission denied' error
}

func TestUndeleteMovesFilesToCorrectDestination(t *testing.T) {
	// Arrange
	is := is.New(t)

	baseFS := afero.NewOsFs()
	readonlyBase := afero.NewReadOnlyFs(baseFS)
	testFS := afero.NewCopyOnWriteFs(readonlyBase, afero.NewMemMapFs())

	// For the purposes of testing, we need to touch the test files so that they copy through to the in-memory overlay.
	// This is due to the following:
	// https://github.com/spf13/afero#copyonwritefs
	// > Removing and Renaming files present only in the base layer is not currently permitted.
	// > If a file is present in the base layer and the overlay, only the overlay will be removed/renamed.

	for _, testFile := range []string{
		"testdata/populated/Deleted Games and Apps/Bugsnax_20210717131548.jpg",
		"testdata/populated/Deleted Games and Apps/Control Ultimate Edition_20210709205443.jpg",
		"testdata/populated/Deleted Games and Apps/DEATH STRANDING DIRECTOR'S CUT_20210927062750.jpg",
	} {
		is.NoErr(afero.WriteFile(testFS, testFile, []byte("bump"), 0o644)) // could not bump test files in overlay FS
	}

	// Act
	org, err := undelete.New(
		undelete.Filesystem(testFS),
		undelete.Path("testdata/populated/Deleted Games and Apps"),
	)
	is.NoErr(err)
	is.NoErr(org.Prepare())
	is.NoErr(org.Undelete())

	// Assert
	_, err = testFS.Stat("testdata/populated/Bugsnax/Bugsnax_20210717131548.jpg")
	is.NoErr(err) // image not at expected path
	_, err = testFS.Stat("testdata/populated/Control Ultimate Edition/Control Ultimate Edition_20210709205443.jpg")
	is.NoErr(err) // image not at expected path
	_, err = testFS.Stat("testdata/populated/DEATH STRANDING DIRECTOR'S CUT/" +
		"DEATH STRANDING DIRECTOR'S CUT_20210927062750.jpg")
	is.NoErr(err) // image not at expected path
}

func TestSiblingsSliceUnaffected(t *testing.T) {
	// Arrange
	is := is.New(t)

	org, err := undelete.New(undelete.Path("testdata/populated/Deleted Games and Apps"))
	is.NoErr(err)

	// Act
	names1 := org.GetNames()
	is.True(len(names1) >= 1)

	names1[0] = "my test string"

	names2 := org.GetNames()
	is.True(len(names2) >= 1)

	// Assert
	is.True(names2[0] != "my test string")
}

func TestUndeleteWithoutPrepareFails(t *testing.T) {
	// Arrange
	is := is.New(t)

	org, err := undelete.New(undelete.Path("testdata/populated/Deleted Games and Apps"))
	is.NoErr(err)

	// Act
	err = org.Undelete()

	// Assert
	is.True(err != nil) // error should not be nil

	var target *os.LinkError

	is.True(errors.As(err, &target)) // error should be of type *os.LinkError
	is.Equal(target.Op, "rename")    // error operation should be 'rename'
}
