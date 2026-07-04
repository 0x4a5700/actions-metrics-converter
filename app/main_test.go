package main

import (
	"testing"

	"github.com/0x4a5700/actions-metrics-converter/internal/apperr"
	"github.com/stretchr/testify/assert"
)

func TestMissingGitHubSecretEnvVar(t *testing.T) {
	cases := []struct {
		name      string
		envVarVal string
		wantErr   bool
	}{
		{name: "empty string", envVarVal: "", wantErr: true},
		{name: "secret present", envVarVal: "s3cret", wantErr: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GITHUB_WEBHOOK_SECRET", tc.envVarVal)

			secret, err := getSecretBytes()
			if tc.wantErr {
				assert.ErrorAs(t, err, &apperr.MissingEnvVar{})
			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte(tc.envVarVal), secret)
			}
		})
	}
}
