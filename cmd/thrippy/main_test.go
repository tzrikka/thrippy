package main

import (
	"path/filepath"
	"testing"
)

func TestConfigDirAndFile(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", d)

	got := configFile(configDir())
	want := filepath.Join(d, configDirName, configFileName)
	if got.SourceURI() != want {
		t.Errorf("configFile(configDir()) = %q, want %q", got.SourceURI(), want)
	}
}
