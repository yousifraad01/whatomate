package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shridarpatil/whatomate/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The default config file is optional (containers may be configured through
// environment variables only), but an explicitly named file must exist.
func TestLoad_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("WHATOMATE_SERVER__PORT", "9191")
	cfg, err := config.Load(config.DefaultConfigPath)
	require.NoError(t, err, "missing default config file is tolerated")
	assert.Equal(t, 9191, cfg.Server.Port)

	_, err = config.Load(filepath.Join(dir, "explicit-missing.toml"))
	require.Error(t, err, "an explicitly requested file must exist")
}
