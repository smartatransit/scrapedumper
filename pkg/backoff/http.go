package backoff

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer           Doer
	retries        int
	factor         int
	backoffSeconds int
	jitter         bool
}

func CreateBackoffClientWithDefaults(doer Doer) Client {
	return Client{
		doer:           doer,
		retries:        3,
		factor:         2,
		backoffSeconds: 10,
		jitter:         true,
	}
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	var attempt int
	for {
		resp, err := c.doer.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			if attempt >= c.retries {
				return resp, nil

			} else if IsRetryable(resp.StatusCode) {
				attempt++
				time.Sleep(c.CalculateBackoff(attempt))
				continue
			}
		}
		return resp, nil
	}
}

func (c Client) CalculateBackoff(attempt int) time.Duration {
	// take our backoff seconds, and multiply them by factor^attempt
	exponentialBackoff := float64(c.backoffSeconds) * math.Pow(float64(c.factor), float64(attempt))
	if c.jitter {
		//if jitter, take half of our backoff, and add a random amount of time to it, so at least we'd have half, at most the full backoff
		exponentialBackoff = (exponentialBackoff / 2) + ((exponentialBackoff / 2) * rand.Float64())
	}
	dur, _ := time.ParseDuration(fmt.Sprintf("%ss", exponentialBackoff))
	return dur
}

func IsRetryable(status int) bool {
	switch {
	case status == http.StatusOK:
		return true
	case status == http.StatusTooManyRequests:
		return true
	case status >= 500:
		return true
	case status == http.StatusBadRequest:
		return false
	default:
		return false
	}
}
