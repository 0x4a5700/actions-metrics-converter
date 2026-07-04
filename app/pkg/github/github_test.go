package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalAllPayloads(t *testing.T) {
	path := "testdata/wf_*.json"
	files, err := filepath.Glob(path)
	require.NoError(t, err)
	require.NotEmpty(t, files, "no files found matching %s", path)
	for _, f := range files {
		data, err := os.ReadFile(f)
		require.NoError(t, err, "read %s", f)
		var p WorkflowJobPayload
		assert.NoError(t, json.Unmarshal(data, &p), "unmarshal %s", filepath.Base(f))
	}
}
