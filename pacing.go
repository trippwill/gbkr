package gbkr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// NoPacing returns nil, which disables pacing when passed to [WithRateLimit].
func NoPacing() *PacingPolicy { return nil }

// pacingRule describes a single endpoint pacing constraint.
type pacingRule struct {
	pathPrefix    string
	method        string // empty = any method
	isRateLimit   bool   // true = rate limiter, false = semaphore
	interval      time.Duration
	burst         int
	maxConcurrent int
}

// ibkrPacingTable lists IBKR pacing rules sorted by path length descending
// (longest prefix first) so the first prefix match is the most specific.
var ibkrPacingTable = []pacingRule{
	// /iserver/account (longest paths first)
	{pathPrefix: "/iserver/account/pnl/partitioned", method: "GET", isRateLimit: true, interval: 5 * time.Second, burst: 1},
	{pathPrefix: "/iserver/account/orders", method: "GET", isRateLimit: true, interval: 5 * time.Second, burst: 1},
	{pathPrefix: "/iserver/account/trades", method: "GET", isRateLimit: true, interval: 5 * time.Second, burst: 1},

	// /iserver/marketdata
	{pathPrefix: "/iserver/marketdata/snapshot", method: "GET", isRateLimit: true, interval: 100 * time.Millisecond, burst: 10},
	{pathPrefix: "/iserver/marketdata/history", method: "GET", isRateLimit: false, maxConcurrent: 5},

	// /iserver/scanner
	{pathPrefix: "/iserver/scanner/params", method: "GET", isRateLimit: true, interval: 15 * time.Minute, burst: 1},
	{pathPrefix: "/iserver/scanner/run", method: "POST", isRateLimit: true, interval: time.Second, burst: 1},

	// /fyi (longest paths first)
	{pathPrefix: "/fyi/deliveryoptions/device", method: "POST", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/deliveryoptions/email", method: "PUT", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/deliveryoptions/", method: "DELETE", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/deliveryoptions", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/notifications/more", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/notifications/", method: "PUT", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/notifications", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/unreadnumber", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/disclaimer/", method: "PUT", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/disclaimer/", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/settings/", method: "POST", isRateLimit: true, interval: time.Second, burst: 1},
	{pathPrefix: "/fyi/settings", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},

	// /pa
	{pathPrefix: "/pa/transactions", method: "POST", isRateLimit: true, interval: 15 * time.Minute, burst: 1},
	{pathPrefix: "/pa/performance", method: "POST", isRateLimit: true, interval: 15 * time.Minute, burst: 1},
	{pathPrefix: "/pa/summary", method: "POST", isRateLimit: true, interval: 15 * time.Minute, burst: 1},

	// /portfolio
	{pathPrefix: "/portfolio/subaccounts", method: "GET", isRateLimit: true, interval: 5 * time.Second, burst: 1},
	{pathPrefix: "/portfolio/accounts", method: "GET", isRateLimit: true, interval: 5 * time.Second, burst: 1},

	// /sso
	{pathPrefix: "/sso/validate", method: "GET", isRateLimit: true, interval: time.Minute, burst: 1},

	// /tickle
	{pathPrefix: "/tickle", method: "GET", isRateLimit: true, interval: time.Second, burst: 1},
}

// pathLimiter holds the runtime limiter for a single pacing rule.
type pathLimiter struct {
	prefix      string
	method      string
	rateLimiter *rate.Limiter // non-nil for rate-limit rules
	semaphore   chan struct{} // non-nil for semaphore rules
}

// PacingObserver receives notifications about pacing events.
type PacingObserver interface {
	OnPacingWait(path string, waitDuration time.Duration)
	OnCacheHit(path string, age time.Duration)
	OnCacheMiss(path string)
}

// PacingPolicy enforces IBKR API rate limits and concurrency constraints.
type PacingPolicy struct {
	global   *rate.Limiter
	rules    []pathLimiter
	observer PacingObserver
}

// newDefaultPacingPolicy builds a PacingPolicy from the ibkrPacingTable.
func newDefaultPacingPolicy() *PacingPolicy {
	rules := make([]pathLimiter, 0, len(ibkrPacingTable))
	for _, r := range ibkrPacingTable {
		pl := pathLimiter{
			prefix: r.pathPrefix,
			method: r.method,
		}
		if r.isRateLimit {
			pl.rateLimiter = rate.NewLimiter(rate.Every(r.interval), r.burst)
		} else {
			pl.semaphore = make(chan struct{}, r.maxConcurrent)
		}
		rules = append(rules, pl)
	}
	return &PacingPolicy{
		global: rate.NewLimiter(rate.Limit(10), 3),
		rules:  rules,
	}
}

// waitForSlot blocks until the pacing constraint for path is satisfied.
// Returns an error wrapping ErrPacingWait if the context is cancelled.
func (p *PacingPolicy) waitForSlot(ctx context.Context, method, path string) error {
	for i := range p.rules {
		r := &p.rules[i]
		if !strings.HasPrefix(path, r.prefix) {
			continue
		}
		if r.method != "" && r.method != method {
			continue
		}

		start := time.Now()

		if r.rateLimiter != nil {
			if err := r.rateLimiter.Wait(ctx); err != nil {
				return fmt.Errorf("%w: %w", ErrPacingWait, err)
			}
			if p.observer != nil {
				p.observer.OnPacingWait(path, time.Since(start))
			}
			return nil
		}

		// Semaphore path.
		select {
		case r.semaphore <- struct{}{}:
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrPacingWait, ctx.Err())
		}
		if p.observer != nil {
			p.observer.OnPacingWait(path, time.Since(start))
		}
		return nil
	}

	// No specific rule matched; use the global limiter.
	start := time.Now()
	if err := p.global.Wait(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrPacingWait, err)
	}
	if p.observer != nil {
		p.observer.OnPacingWait(path, time.Since(start))
	}
	return nil
}

// releaseSlot releases a semaphore slot for the given path.
// No-op for rate-limited endpoints.
func (p *PacingPolicy) releaseSlot(method, path string) {
	for i := range p.rules {
		r := &p.rules[i]
		if !strings.HasPrefix(path, r.prefix) {
			continue
		}
		if r.method != "" && r.method != method {
			continue
		}
		if r.semaphore != nil {
			<-r.semaphore
		}
		return
	}
}
