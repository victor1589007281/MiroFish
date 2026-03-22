package simulation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExperts(t *testing.T) {
	content := `experts:
  - name: test_optimist
    system_prompt: "You are optimistic."
  - name: test_pessimist
    system_prompt: "You are pessimistic."
`
	dir := t.TempDir()
	path := filepath.Join(dir, "experts.yaml")
	os.WriteFile(path, []byte(content), 0644)

	experts, err := LoadExperts(path)
	require.NoError(t, err)
	assert.Len(t, experts, 2)
	assert.Equal(t, "test_optimist", experts[0].Name)
	assert.Equal(t, "You are optimistic.", experts[0].SystemPrompt)
}

func TestLoadExperts_FileNotFound(t *testing.T) {
	_, err := LoadExperts("/nonexistent/path.yaml")
	assert.Error(t, err)
}

func TestLoadExpertsOrDefault(t *testing.T) {
	experts := LoadExpertsOrDefault("/nonexistent/path.yaml")
	assert.Len(t, experts, 5) // should fall back to defaults
	assert.Equal(t, "optimist", experts[0].Name)
}
