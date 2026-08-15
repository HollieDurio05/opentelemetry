package helper

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/HollieDurio05/opentelemetry/pkg/stanza/entry"
)

type Parser struct {
	ErrorLimiter *RateLimiter
}

type RateLimiter struct {
	mu       sync.Mutex
	lastLog  time.Time
	interval time.Duration
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if now.Sub(rl.lastLog) >= rl.interval {
		rl.lastLog = now
		return true
	}
	return false
}

func (p *Parser) HandleError(ctx context.Context, err error, e *entry.Entry) {
	if p.ErrorLimiter == nil {
		p.ErrorLimiter = NewRateLimiter(time.Second)
	}

	// Check if the source path matches the collector's own log file destination
	isCollectorLog := false
	if e != nil {
		if path, ok := e.Attributes["log.file.path"].(string); ok {
			if filepath.Base(path) == "otelcol.log" {
				isCollectorLog = true
			}
		}
	}

	if isCollectorLog {
		// Suppress completely to prevent recursive logging loops
		return
	}

	if p.ErrorLimiter.Allow() {
		// Emit a non-looping warning/error log
		fmt.Printf("Parser warning: %v\n", err)
	}
}
