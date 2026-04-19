package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase/odds"
	"github.com/google/uuid"
)

// OddsCache implements odds caching using Redis
type OddsCache struct {
	redis *RedisCache
}

func NewOddsCache(redis *RedisCache) *OddsCache {
	return &OddsCache{redis: redis}
}

// CacheMatchOdds stores match odds in cache
func (c *OddsCache) CacheMatchOdds(ctx context.Context, match *domain.Match) error {
	key := fmt.Sprintf("match_odds:%s", match.ID)
	return c.redis.Set(ctx, key, match, 30*time.Second)
}

// GetMatchOdds retrieves match odds from cache
func (c *OddsCache) GetMatchOdds(ctx context.Context, matchID string) (*domain.Match, error) {
	key := fmt.Sprintf("match_odds:%s", matchID)

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("match odds not found in cache: %w", err)
	}

	var match domain.Match
	if err := json.Unmarshal(fmt.Appendf(nil, "%v", data), &match); err != nil {
		return nil, fmt.Errorf("failed to unmarshal match odds: %w", err)
	}

	return &match, nil
}

// CacheLiveMatches stores live matches list
func (c *OddsCache) CacheLiveMatches(ctx context.Context, matches []*domain.Match) error {
	key := "live_matches"
	return c.redis.Set(ctx, key, matches, 10*time.Second)
}

// GetLiveMatches retrieves live matches from cache
func (c *OddsCache) GetLiveMatches(ctx context.Context) ([]*domain.Match, error) {
	key := "live_matches"

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("live matches not found in cache: %w", err)
	}

	var matches []*domain.Match
	if err := json.Unmarshal(fmt.Appendf(nil, "%v", data), &matches); err != nil {
		return nil, fmt.Errorf("failed to unmarshal live matches: %w", err)
	}

	return matches, nil
}

// CacheCalculatedOdds stores calculated odds for a bet
func (c *OddsCache) CacheCalculatedOdds(ctx context.Context, betID uuid.UUID, result *odds.BetOddsResult) error {
	key := fmt.Sprintf("calculated_odds:%s", betID.String())
	return c.redis.Set(ctx, key, result, 5*time.Minute)
}

// GetCalculatedOdds retrieves calculated odds from cache
func (c *OddsCache) GetCalculatedOdds(ctx context.Context, betID uuid.UUID) (*odds.BetOddsResult, error) {
	key := fmt.Sprintf("calculated_odds:%s", betID.String())

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("calculated odds not found in cache: %w", err)
	}

	var result odds.BetOddsResult
	if err := json.Unmarshal(fmt.Appendf(nil, "%v", data), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal calculated odds: %w", err)
	}

	return &result, nil
}

// InvalidateMatchOdds removes match odds from cache
func (c *OddsCache) InvalidateMatchOdds(ctx context.Context, matchID string) error {
	key := fmt.Sprintf("match_odds:%s", matchID)
	return c.redis.Delete(ctx, key)
}

// InvalidateLiveMatches removes live matches from cache
func (c *OddsCache) InvalidateLiveMatches(ctx context.Context) error {
	key := "live_matches"
	return c.redis.Delete(ctx, key)
}
