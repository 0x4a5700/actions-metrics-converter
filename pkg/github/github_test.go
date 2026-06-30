package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnmarshalAllPayloads(t *testing.T) {
	files, err := filepath.Glob(".testdata/wf_*.json")
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
