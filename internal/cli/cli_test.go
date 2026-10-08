package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunHelpAndVersion(t *testing.T) {
	var output bytes.Buffer
	if code := Run(context.Background(), []string{"--help"}, &output, &output); code != 0 || !strings.Contains(output.String(), "verify npm") {
		t.Fatalf("help: code=%d output=%q", code, output.String())
	}

	output.Reset()
	if code := Run(context.Background(), []string{"--version"}, &output, &output); code != 0 || !strings.Contains(output.String(), "0.1.0-dev") {
		t.Fatalf("version: code=%d output=%q", code, output.String())
	}
}

func TestRunRejectsInvalidUsage(t *testing.T) {
	for _, args := range [][]string{
		{"unknown"},
		{"verify"},
		{"verify", "npm", "--json", "--sarif", "demo"},
		{"verify", "npm", "--timeout", "0s", "demo"},
		{"verify", "wat", "demo"},
	} {
		var stderr bytes.Buffer
		if code := Run(context.Background(), args, &bytes.Buffer{}, &stderr); code != 2 {
			t.Fatalf("args %v: got code %d stderr %q", args, code, stderr.String())
		}
	}
}

func TestRunRequiresPackageName(t *testing.T) {
	var stderr bytes.Buffer
	if code := Run(context.Background(), []string{"verify", "npm"}, &bytes.Buffer{}, &stderr); code != 2 {
		t.Fatalf("got code %d stderr %q", code, stderr.String())
	}
}
