package dedupe_test

import (
	"testing"

	"github.com/matryer/is"

	"go.jlucktay.dev/playstation-share-dedupe/pkg/dedupe"
)

func TestLifecycle(t *testing.T) {
	is := is.New(t)

	d, err := dedupe.New("directory")
	is.NoErr(err)
	is.NoErr(d.Close())
}
