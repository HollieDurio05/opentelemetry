package json

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HollieDurio05/opentelemetry/pkg/stanza/entry"
	"github.com/HollieDurio05/opentelemetry/pkg/stanza/operator/helper"
)

type JSONParser struct {
	helper.Parser
}

func NewJSONParser() *JSONParser {
	return &JSONParser{
		Parser: helper.Parser{
			ErrorLimiter: helper.NewRateLimiter(time.Second),
		},
	}
}

func (p *JSONParser) Process(ctx context.Context, e *entry.Entry) error {
	var bodyStr string
	switch b := e.Body.(type) {
	case string:
		bodyStr = b
	case []byte:
		bodyStr = string(b)
	default:
		bodyStr = fmt.Sprintf("%v", e.Body)
	}

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(bodyStr), &parsed)
	if err != nil {
		// If parsing fails, the raw log line must be preserved in the Body of the LogRecord
		// rather than dropping the log entirely, and a non-looping warning should be emitted.
		p.HandleError(ctx, err, e)
		// Keep the raw body as is, do not return error to drop it
		return nil
	}

	// Parse fields: ts, level, msg, caller, stacktrace
	// ts (float64 epoch or ISO8601 string) -> Timestamp
	if tsVal, ok := parsed["ts"]; ok {
		if t, err := parseTimestamp(tsVal); err == nil {
			e.Timestamp = t
		}
	}

	// level (string) -> SeverityText / SeverityNumber (we can store them in attributes or map them)
	if levelVal, ok := parsed["level"]; ok {
		if levelStr, ok := levelVal.(string); ok {
			e.Attributes["level"] = levelStr
		}
	}

	// msg (string) -> Body
	if msgVal, ok := parsed["msg"]; ok {
		if msgStr, ok := msgVal.(string); ok {
			e.Body = msgStr
		}
	}

	// caller (string) -> Attributes (log.logger.name or custom attribute)
	if callerVal, ok := parsed["caller"]; ok {
		if callerStr, ok := callerVal.(string); ok {
			e.Attributes["log.logger.name"] = callerStr
		}
	}

	// stacktrace (string) -> Attributes (exception.stacktrace)
	if stackVal, ok := parsed["stacktrace"]; ok {
		if stackStr, ok := stackVal.(string); ok {
			e.Attributes["exception.stacktrace"] = stackStr
		}
	}

	// Map other fields to attributes
	for k, v := range parsed {
		if k != "ts" && k != "level" && k != "msg" && k != "caller" && k != "stacktrace" {
			e.Attributes[k] = v
		}
	}

	return nil
}

func parseTimestamp(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case float64:
		// float64 epoch seconds (e.g., 1672531199.123456)
		sec := int64(val)
		nsec := int64((val - float64(sec)) * 1e9)
		return time.Unix(sec, nsec).UTC(), nil
	case string:
		// ISO8601 string
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.999999999Z07:00",
			"2006-01-02 15:04:05.999999999",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, val); err == nil {
				return t.UTC(), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format: %v", v)
}
