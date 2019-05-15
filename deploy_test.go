package main

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFsFileExists(t *testing.T) {
	ok, err := fs.FileExists("testdata/ParameterFile1.json")
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = fs.FileExists("testdata/notexist")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestFsExpanduser(t *testing.T) {
	old := fs
	defer func() { fs = old }()

	tests := []struct {
		Homedir string
		Path    string
		Out     string
	}{
		{"/home/example", "foo", "foo"},
		{"/home/example", "~/foo", "/home/example/foo"},
		{"/home/example", "/foo", "/foo"},
		{"/home/example", "C:/foo", "C:/foo"},
		{"/home/example", "C:\\foo", "C:\\foo"},
		{"C:/Users/foo", "~/foo", "C:/Users/foo/foo"},
	}

	for _, test := range tests {
		t.Run(test.Path, func(t *testing.T) {
			fs.UserHomeDir = func() (dir string, err error) {
				return test.Homedir, nil
			}

			out, err := fs.ExpandUser(test.Path)
			require.NoError(t, err)
			require.Equal(t, test.Out, out)
		})
	}
}

func TestFindManifest(t *testing.T) {
	old := fs
	defer func() { fs = old }()

	tests := []struct {
		Cwd      string
		Manifest string
		Expect   string
		Ok       bool
	}{
		{"/home/bob", "/home/bob/.cftool.yml", ".cftool.yml", true},
		{"C:/home/bob", "C:/home/bob/.cftool.yml", ".cftool.yml", true},
		{"/home/bob", "", ".cftool.yml", false},
		{"C:/home/bob", "", ".cftool.yml", false},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			fs = old

			fs.Getwd = func() (cwd string, err error) {
				return test.Cwd, nil
			}

			fs.FileExists = func(path string) (ok bool, err error) {
				return test.Manifest != "" && test.Manifest == path, nil
			}

			if strings.HasPrefix(test.Cwd, "C:") {
				fs.VolumeName = func(path string) (out string) {
					return path[:2]
				}
			}

			out, err := findManifest()

			if test.Ok {
				require.NoError(t, err)
				require.Equal(t, test.Expect, out)
			} else {
				require.Error(t, err)
				require.Empty(t, out)
			}
		})
	}
}
