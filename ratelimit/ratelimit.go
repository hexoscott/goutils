/*
package ratelimit is a really nice implementation of a rate limiter suitable for production
use.  It brings the concept of having windowed limits e.g. 5 per second, 10 per hour and handling
the wait balancing off both of these limits

taken from the excellent concurrency in Go book written by Katherine Cox-Buday
*/
package ratelimit

import (
	"context"
	"sort"
	"time"

	"golang.org/x/time/rate"
)

// a RateLimiter will wait until the bucket has a token available to service
// the request
type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

type multilimiter struct {
	limiters []RateLimiter
}

// Per is a utility function to give a rate.Limit value for X events in Y timeframe
func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

// MultiLimiter takes in a series of RateLimiters, sorts them by their limit ascending and
// returns a multilimiter
func MultiLimiter(limiters ...RateLimiter) *multilimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}

	sort.Slice(limiters, byLimit)

	return &multilimiter{limiters: limiters}
}

// Wait iterates through the limiters until all buckets have a token available to serve a request
func (l *multilimiter) Wait(ctx context.Context) error {
	for _, x := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Limit returns the most restrictive limit from the collection of limiters
func (l *multilimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}
