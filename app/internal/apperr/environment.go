package apperr

import (
	"fmt"
	"strings"
)

type MissingEnvVar struct {
	VarName string
	Fatal   bool
}

func (e MissingEnvVar) Error() string {
	s := strings.Builder{}
	_, _ = fmt.Fprintf(&s, "Missing environment variable %q\n", e.VarName)
	if e.Fatal {
		s.WriteString(" which is required to continue")
	}
	return s.String()
}
