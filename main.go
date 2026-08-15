package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/HollieDurio05/opentelemetry/pkg/stanza/adapter"
	"github.com/HollieDurio05/opentelemetry/pkg/stanza/entry"
	"github.com/HollieDurio05/opentelemetry/pkg/stanza/operator/parser/json"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func main() {
	fmt.Println("Starting OpenTelemetry Filelog Receiver Simulation...")

	// Create a temporary log file
	tmpDir, err := os.MkdirTemp("", "otelcol-logs")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	logFilePath := filepath.Join(tmpDir, "otelcol.log")
	
	// Write sample logs (valid Zap JSON and invalid JSON)
	logs := []string{
		`{"level":"info","ts":1609459200.123456,"caller":"service/service.go:120","msg":"Starting OpenTelemetry Collector","version":"v0.90.0"}`,
		`{"level":"warn","ts":1609459201.789012,"caller":"service/service.go:150","msg":"Configuration warning","warning_code":42}`,
		`{"level":"error","ts":1609459202.345678,"caller":"service/service.go:200","msg":"Failed to connect to backend","stacktrace":"main.go:50\nmain.go:100"}`,
		`invalid json line that should not cause infinite loop`,
	}

	err = os.WriteFile(logFilePath, []byte(""), 0644)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	for _, logLine := range logs {
		_, err = f.WriteString(logLine + "\n")
		if err != nil {
			fmt.Printf("Failed to write log line: %v\n", err)
			return
	}
	}

	// Simulate reading and parsing
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		fmt.Printf("Failed to read log file: %v\n", err)
		return
	}

	// Split content by newline
	var lines []string
	var currentLine []byte
	for _, b := range content {
		if b == '\n' {
			if len(currentLine) > 0 {
				lines = append(lines, string(currentLine))
				currentLine = nil
			}
		} else {
			currentLine = append(currentLine, b)
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}

	parser := json.NewJSONParser()
	ctx := context.Background()

	for i, line := range lines {
		fmt.Printf("\n--- Processing Log Line %d ---\n", i+1)
		fmt.Printf("Raw: %s\n", line)

		e := entry.New()
		e.Attributes["log.file.path"] = logFilePath
		e.Body = line

		err := parser.Process(ctx, e)
		if err != nil {
			fmt.Printf("Parser error: %v\n", err)
			continue
		}

		// Convert to plog.LogRecord
		lr := adapter.Convert(e)

		// Print mapped fields
		fmt.Printf("Mapped Timestamp: %v\n", lr.Timestamp().AsTime().Format(time.RFC3339Nano))
		fmt.Printf("Mapped SeverityText: %s\n", lr.SeverityText())
		fmt.Printf("Mapped SeverityNumber: %s\n", lr.SeverityNumber().String())
		fmt.Printf("Mapped Body: %s\n", lr.Body().AsString())
		
		fmt.Println("Mapped Attributes:")
		lr.Attributes().Range(func(k string, v pcommon.Value) bool {
			fmt.Printf("  %s: %v\n", k, v.AsRaw())
			return true
		})
	}
}
