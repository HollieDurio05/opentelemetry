package adapter

import (
	"fmt"

	"github.com/HollieDurio05/opentelemetry/pkg/stanza/entry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func Convert(e *entry.Entry) plog.LogRecord {
	lr := plog.NewLogRecord()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(e.Timestamp))

	// Map Body
	if e.Body != nil {
		switch b := e.Body.(type) {
		case string:
			lr.Body().SetStr(b)
		default:
			lr.Body().SetStr(fmt.Sprintf("%v", b))
		}
	}

	// Map Severity
	if levelVal, ok := e.Attributes["level"]; ok {
		if levelStr, ok := levelVal.(string); ok {
			lr.SetSeverityText(levelStr)
			lr.SetSeverityNumber(severityNumber(levelStr))
			delete(e.Attributes, "level")
		}
	}

	// Map Attributes
	for k, v := range e.Attributes {
		switch val := v.(type) {
		case string:
			lr.Attributes().PutStr(k, val)
		case bool:
			lr.Attributes().PutBool(k, val)
		case int64:
			lr.Attributes().PutInt(k, val)
		case float64:
			lr.Attributes().PutDouble(k, val)
		default:
			lr.Attributes().PutStr(k, fmt.Sprintf("%v", val))
		}
	}

	return lr
}

func severityNumber(level string) plog.SeverityNumber {
	switch level {
	case "trace", "TRACE":
		return plog.SeverityNumberTrace
	case "debug", "DEBUG":
		return plog.SeverityNumberDebug
	case "info", "INFO":
		return plog.SeverityNumberInfo
	case "warn", "WARN", "warning", "WARNING":
		return plog.SeverityNumberWarn
	case "error", "ERROR":
		return plog.SeverityNumberError
	case "fatal", "FATAL":
		return plog.SeverityNumberFatal
	default:
		return plog.SeverityNumberUnspecified
	}
}
