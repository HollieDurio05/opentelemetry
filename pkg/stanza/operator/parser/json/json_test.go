package json

import (
	"context"
	"testing"
	"time"

	"github.com/HollieDurio05/opentelemetry/pkg/stanza/entry"
)

func TestJSONParser(t *testing.T) {
	parser := NewJSONParser()
	e := entry.New()
	e.Body = `{"level":"info","ts":1609459200.123456,"caller":"service/service.go:120","msg":"Starting OpenTelemetry Collector","version":"v0.90.0"}`

	err := parser.Process(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert Timestamp matches 2021-01-01T00:00:00.123456Z
	expectedTime := time.Date(2021, 1, 1, 0, 0, 0, 123456000, time.UTC)
	if !e.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, e.Timestamp)
	}

	// Assert SeverityText (level) is info
	if level, ok := e.Attributes["level"].(string); !ok || level != "info" {
		t.Errorf("expected level to be 'info', got %v", e.Attributes["level"])
	}

	// Assert Body is Starting OpenTelemetry Collector
	if body, ok := e.Body.(string); !ok || body != "Starting OpenTelemetry Collector" {
		t.Errorf("expected body to be 'Starting OpenTelemetry Collector', got %v", e.Body)
	}

	// Assert caller is mapped to attributes (log.logger.name)
	if caller, ok := e.Attributes["log.logger.name"].(string); !ok || caller != "service/service.go:120" {
		t.Errorf("expected caller to be 'service/service.go:120', got %v", e.Attributes["log.logger.name"])
	}

	// Assert version is mapped to attributes
	if version, ok := e.Attributes["version"].(string); !ok || version != "v0.90.0" {
		t.Errorf("expected version to be 'v0.90.0', got %v", e.Attributes["version"])
	}
}

func TestJSONParser_InvalidJSON(t *testing.T) {
	parser := NewJSONParser()
	e := entry.New()
	invalidRaw := `{"level":"info", invalid json`
	e.Body = invalidRaw

	err := parser.Process(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Raw log line must be preserved in Body
	if body, ok := e.Body.(string); !ok || body != invalidRaw {
		t.Errorf("expected body to preserve raw line, got %v", e.Body)
	}
}

func TestJSONParser_CollectorLogSuppression(t *testing.T) {
	parser := NewJSONParser()
	e := entry.New()
	e.Attributes["log.file.path"] = "/var/log/otelcol.log"
	e.Body = `invalid json`

	err := parser.Process(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body, ok := e.Body.(string); !ok || body != "invalid json" {
		t.Errorf("expected body to preserve raw line, got %v", e.Body)
	}
}
