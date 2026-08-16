package cache

import "time"

const defaultRecentWeekFreshness = 2 * time.Minute

type FreshnessPolicy struct {
	maxAge time.Duration
}

func DefaultFreshnessPolicy() FreshnessPolicy {
	return FreshnessPolicy{maxAge: defaultRecentWeekFreshness}
}

func (p FreshnessPolicy) Allows(refreshedAt, now time.Time) bool {
	if refreshedAt.IsZero() {
		return false
	}
	if p.maxAge <= 0 {
		p.maxAge = defaultRecentWeekFreshness
	}
	return now.Sub(refreshedAt) < p.maxAge
}
