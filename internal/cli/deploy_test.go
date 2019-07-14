package cli

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFindManifest(t *testing.T) {
	t.Run("no manifest", func(t *testing.T) {
		dirname, err := ioutil.TempDir("", "cftool-test")
		require.NoError(t, err)

		result, err := findManifest(dirname)
		require.Error(t, err)
		require.Equal(t, "", result)
	})

	t.Run("has manifest", func(t *testing.T) {
		dirname, err := ioutil.TempDir("", "cftool-test")
		require.NoError(t, err)
		manifestPath := filepath.Join(dirname, ".cftool.yml")
		require.NoError(t, ioutil.WriteFile(manifestPath, []byte{}, 0777))
		innerDir := filepath.Join(dirname, "inner")
		require.NoError(t, os.Mkdir(innerDir, 0777))

		// find from the same directory as manifest
		result, err := findManifest(dirname)
		require.NoError(t, err)
		require.Equal(t, manifestPath, result)

		// find one directory deeper than the manifest
		result, err = findManifest(innerDir)
		require.NoError(t, err)
		require.Equal(t, manifestPath, result)
	})
}
