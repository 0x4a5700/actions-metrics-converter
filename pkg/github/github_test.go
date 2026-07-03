package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshalAllPayloads(t *testing.T) {
	path := "testdata/wf_*.json"
	files, err := filepath.Glob(path)
	if len(files) < 1 {
		t.Fatalf("no files found matching %s", path)
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var p WorkflowJobPayload
		if err := json.Unmarshal(data, &p); err != nil {
			t.Errorf("FAIL %s: %v", filepath.Base(f), err)
		} else {
			t.Logf("OK   %s", filepath.Base(f))
		}
	}
}
