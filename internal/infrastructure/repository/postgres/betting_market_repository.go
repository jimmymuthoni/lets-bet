package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/betting-platform/internal/core/domain"
)

// BettingMarketRepository implements betting market repository using PostgreSQL
type BettingMarketRepository struct {
	db *sql.DB
}

// NewBettingMarketRepository creates a new betting market repository
func NewBettingMarketRepository(db *sql.DB) *BettingMarketRepository {
	return &BettingMarketRepository{db: db}
}

// Create creates a new betting market
func (r *BettingMarketRepository) Create(ctx context.Context, market *domain.Market) error {
	query := `
		INSERT INTO betting_markets (
			id, market_id, event_id, market_type, market_name, status,
			suspended_at, provider, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var suspendedAt sql.NullTime
	if market.SuspendedAt != nil {
		suspendedAt = sql.NullTime{Time: *market.SuspendedAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		market.ID, market.ID, market.MatchID, string(market.Type), market.Name,
		string(market.Status), suspendedAt, "system", time.Now(), time.Now(),
	)

	return err
}

// GetByID retrieves a betting market by ID
func (r *BettingMarketRepository) GetByID(ctx context.Context, id string) (*domain.Market, error) {
	query := `
		SELECT id, market_id, event_id, market_type, market_name, status,
			   suspended_at, provider, created_at, updated_at
		FROM betting_markets
		WHERE id = $1
	`

	var market domain.Market
	var marketType, name, status, provider string
	var suspendedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&market.ID, &market.MatchID, &marketType, &name, &status,
		&suspendedAt, &provider, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	market.Type = domain.MarketType(marketType)
	market.Name = name
	market.Status = domain.MarketStatus(status)

	if suspendedAt.Valid {
		market.SuspendedAt = &suspendedAt.Time
	}

	return &market, nil
}

// GetByMatch retrieves betting markets by match ID
func (r *BettingMarketRepository) GetByMatch(ctx context.Context, matchID string) ([]*domain.Market, error) {
	query := `
		SELECT id, market_id, event_id, market_type, market_name, status,
			   suspended_at, provider, created_at, updated_at
		FROM betting_markets
		WHERE event_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []*domain.Market
	for rows.Next() {
		var market domain.Market
		var marketType, name, status, provider string
		var suspendedAt sql.NullTime
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&market.ID, &market.MatchID, &marketType, &name, &status,
			&suspendedAt, &provider, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		market.Type = domain.MarketType(marketType)
		market.Name = name
		market.Status = domain.MarketStatus(status)

		if suspendedAt.Valid {
			market.SuspendedAt = &suspendedAt.Time
		}

		markets = append(markets, &market)
	}

	return markets, nil
}

// Update updates a betting market
func (r *BettingMarketRepository) Update(ctx context.Context, market *domain.Market) error {
	query := `
		UPDATE betting_markets
		SET market_type = $2, market_name = $3, status = $4, suspended_at = $5,
			updated_at = $6
		WHERE id = $1
	`

	var suspendedAt sql.NullTime
	if market.SuspendedAt != nil {
		suspendedAt = sql.NullTime{Time: *market.SuspendedAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		market.ID, string(market.Type), market.Name, string(market.Status),
		suspendedAt, time.Now(),
	)

	return err
}

// UpdateStatus updates betting market status
func (r *BettingMarketRepository) UpdateStatus(ctx context.Context, id string, status domain.MarketStatus) error {
	query := `
		UPDATE betting_markets
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, string(status), time.Now())
	return err
}

// Delete deletes a betting market
func (r *BettingMarketRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM betting_markets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetActiveMarkets retrieves all active betting markets
func (r *BettingMarketRepository) GetActiveMarkets(ctx context.Context) ([]*domain.Market, error) {
	query := `
		SELECT id, market_id, event_id, market_type, market_name, status,
			   suspended_at, provider, created_at, updated_at
		FROM betting_markets
		WHERE status = 'OPEN'
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []*domain.Market
	for rows.Next() {
		var market domain.Market
		var marketType, name, status, provider string
		var suspendedAt sql.NullTime
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&market.ID, &market.MatchID, &marketType, &name, &status,
			&suspendedAt, &provider, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		market.Type = domain.MarketType(marketType)
		market.Name = name
		market.Status = domain.MarketStatus(status)

		if suspendedAt.Valid {
			market.SuspendedAt = &suspendedAt.Time
		}

		markets = append(markets, &market)
	}

	return markets, nil
}
