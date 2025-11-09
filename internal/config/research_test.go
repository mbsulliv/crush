package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResearchModeFlag(t *testing.T) {
	t.Parallel()

	testDir := t.TempDir()
	workingDir := filepath.Join(testDir, "project")
	err := os.MkdirAll(workingDir, 0o755)
	require.NoError(t, err)

	t.Run("research flag sets mode to research", func(t *testing.T) {
		cfg, err := Load(workingDir, "", false, true)
		require.NoError(t, err)
		assert.Equal(t, "research", cfg.Options.Mode)
	})

	t.Run("no research flag defaults to coder mode", func(t *testing.T) {
		cfg, err := Load(workingDir, "", false, false)
		require.NoError(t, err)
		assert.Equal(t, "coder", cfg.Options.Mode)
	})

	t.Run("research flag overrides config file", func(t *testing.T) {
		// Create a config file with mode set to coder
		configPath := filepath.Join(workingDir, "crush.json")
		configData := `{"options": {"mode": "coder"}}`
		err := os.WriteFile(configPath, []byte(configData), 0o644)
		require.NoError(t, err)

		// Load with research flag should override
		cfg, err := Load(workingDir, "", false, true)
		require.NoError(t, err)
		assert.Equal(t, "research", cfg.Options.Mode)
	})

	t.Run("existing research mode in config preserved without flag", func(t *testing.T) {
		// Create a new temp dir for this test
		testDir2 := t.TempDir()
		workingDir2 := filepath.Join(testDir2, "project")
		err := os.MkdirAll(workingDir2, 0o755)
		require.NoError(t, err)

		// Create a config file with mode set to research
		configPath := filepath.Join(workingDir2, "crush.json")
		configData := `{"options": {"mode": "research"}}`
		err = os.WriteFile(configPath, []byte(configData), 0o644)
		require.NoError(t, err)

		// Load without research flag should keep research mode from config
		cfg, err := Load(workingDir2, "", false, false)
		require.NoError(t, err)
		assert.Equal(t, "research", cfg.Options.Mode)
	})
}
