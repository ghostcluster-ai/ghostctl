package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ghostcluster-ai/ghostctl/internal/metadata"
)

func TestDisplayStatusWithoutMetadata(t *testing.T) {
	output := captureStdout(t, func() {
		displayStatus("pr-789", nil, "ghostcluster", "/tmp/kubeconfig.yaml", "not found", false, false)
	})

	if !strings.Contains(output, "Created: unknown") {
		t.Fatalf("expected Created: unknown, got: %s", output)
	}
	if !strings.Contains(output, "TTL: unknown") {
		t.Fatalf("expected TTL: unknown, got: %s", output)
	}
	if !strings.Contains(output, "vCluster not found") {
		t.Fatalf("expected vCluster not found message, got: %s", output)
	}
}

func TestDisplayStatusWithMetadata(t *testing.T) {
	meta := &metadata.ClusterMetadata{
		Name:      "pr-101",
		Namespace: "ghostcluster",
		CreatedAt: time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC),
		TTL:       "1h",
	}

	output := captureStdout(t, func() {
		displayStatus("pr-101", meta, "ghostcluster", "/tmp/kubeconfig.yaml", "running", true, true)
	})

	if !strings.Contains(output, "Created: 2026-02-01 10:30:00") {
		t.Fatalf("expected created time in output, got: %s", output)
	}
	if !strings.Contains(output, "TTL: 1h") {
		t.Fatalf("expected TTL in output, got: %s", output)
	}
	if !strings.Contains(output, "vCluster is accessible") {
		t.Fatalf("expected accessible message, got: %s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	fn()

	_ = writer.Close()
	os.Stdout = originalStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()

	return buf.String()
}
