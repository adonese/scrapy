package scrapers

import (
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"

var (
	randMu sync.Mutex
	// randSource is a shared pseudo-random generator used to rotate user agents
	// and compute jitter. The global functions in math/rand are safe for
	// concurrent use, but we keep an explicit instance to control the seed and
	// remove the dependency on the package-level lock.
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// EffectiveTimeout resolves the timeout to use for outbound HTTP requests.
func (c Config) EffectiveTimeout() time.Duration {
	if c.Timeout > 0 {
		return time.Duration(c.Timeout) * time.Second
	}
	return 30 * time.Second
}

// EffectiveMaxRetries returns the configured retry attempts or a sane default.
func (c Config) EffectiveMaxRetries() int {
	if c.MaxRetries > 0 {
		return c.MaxRetries
	}
	return 3
}

// EffectiveUserAgent picks a random User-Agent string for the request.
func (c Config) EffectiveUserAgent() string {
	if len(c.UserAgents) > 0 {
		idx := randomIndex(len(c.UserAgents))
		if ua := strings.TrimSpace(c.UserAgents[idx]); ua != "" {
			return ua
		}
		for _, candidate := range c.UserAgents {
			if trimmed := strings.TrimSpace(candidate); trimmed != "" {
				return trimmed
			}
		}
	}

	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		return ua
	}

	return defaultUserAgent
}

// BuildHTTPClient constructs an HTTP client with sensible defaults for
// scraping workloads. It respects custom clients and proxy configuration.
func BuildHTTPClient(c Config) *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	timeout := c.EffectiveTimeout()

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if proxy := strings.TrimSpace(c.ProxyURL); proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	if jar, err := cookiejar.New(nil); err == nil {
		client.Jar = jar
	}

	return client
}

// PrepareRequest creates an HTTP request with headers tuned for scraping.
func PrepareRequest(ctx context.Context, method, url string, body io.Reader, c Config) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.EffectiveUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("DNT", "1")

	if len(c.ExtraHeaders) > 0 {
		for key, value := range c.ExtraHeaders {
			if strings.TrimSpace(key) == "" {
				continue
			}
			req.Header.Set(key, value)
		}
	}

	return req, nil
}

// DelayBetweenRequests waits for a randomised interval before issuing a
// request. The range is defined by MinDelayBetweenRequests and
// MaxDelayBetweenRequests. When both are zero no delay is introduced.
func DelayBetweenRequests(ctx context.Context, c Config) error {
	delay := c.randomRequestDelay()
	return wait(ctx, delay)
}

// WaitRetry sleeps for the duration computed for the given retry attempt.
func WaitRetry(ctx context.Context, c Config, attempt int) error {
	return wait(ctx, RetryDelay(c, attempt))
}

// RetryDelay calculates the delay before the next retry attempt using an
// exponential backoff with jitter strategy.
func RetryDelay(c Config, attempt int) time.Duration {
	base := c.RetryBaseDelay
	if base <= 0 {
		base = 1500 * time.Millisecond
	}

	// Exponential backoff with jitter. The first retry waits for base,
	// subsequent retries double the wait duration.
	wait := base * time.Duration(1<<attempt)

	jitterRange := base / 2
	if jitterRange <= 0 {
		jitterRange = base
	}

	jitter := time.Duration(randInt63n(int64(jitterRange) + 1))
	return wait + jitter
}

// wait sleeps for the requested duration unless the context is cancelled.
func wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// randomRequestDelay returns the jittered delay configured for the scraper.
func (c Config) randomRequestDelay() time.Duration {
	min := c.MinDelayBetweenRequests
	max := c.MaxDelayBetweenRequests

	if min <= 0 && max <= 0 {
		return 0
	}

	if min <= 0 {
		min = 250 * time.Millisecond
	}

	if max <= 0 || max < min {
		max = min
	}

	if max == min {
		return min
	}

	delta := max - min
	return min + time.Duration(randInt63n(int64(delta+1)))
}

func randomIndex(length int) int {
	if length <= 1 {
		return 0
	}
	return int(randInt63n(int64(length)))
}

func randInt63n(n int64) int64 {
	if n <= 0 {
		return 0
	}

	randMu.Lock()
	defer randMu.Unlock()
	return randSource.Int63n(n)
}
