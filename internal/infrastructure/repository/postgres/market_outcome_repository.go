package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// MarketOutcomeRepository implements market outcome repository using PostgreSQL
type MarketOutcomeRepository struct {
	db *sql.DB
}

// NewMarketOutcomeRepository creates a new market outcome repository
func NewMarketOutcomeRepository(db *sql.DB) *MarketOutcomeRepository {
	return &MarketOutcomeRepository{db: db}
}

// Create creates a new market outcome
func (r *MarketOutcomeRepository) Create(ctx context.Context, outcome *domain.Outcome) error {
	query := `
		INSERT INTO market_outcomes (
			id, outcome_id, market_id, outcome_name, odds, price, status,
			settled_at, settlement_factor, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var settledAt sql.NullTime
	if outcome.Status == domain.OutcomeStatusWon || outcome.Status == domain.OutcomeStatusLost {
		settledAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		outcome.ID, outcome.ID, outcome.MarketID, outcome.Name, outcome.Odds,
		outcome.Price, string(outcome.Status), settledAt, 1.0, time.Now(), time.Now(),
	)

	return err
}

// GetByID retrieves a market outcome by ID
func (r *MarketOutcomeRepository) GetByID(ctx context.Context, id string) (*domain.Outcome, error) {
	query := `
		SELECT id, outcome_id, market_id, outcome_name, odds, price, status,
			   settled_at, settlement_factor, created_at, updated_at
		FROM market_outcomes
		WHERE id = $1
	`

	var outcome domain.Outcome
	var name, status string
	var settledAt sql.NullTime
	var settlementFactor float64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&outcome.ID, &outcome.MarketID, &name, &outcome.Odds, &outcome.Price,
		&status, &settledAt, &settlementFactor, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, err
	}

	outcome.Name = name
	outcome.Status = domain.OutcomeStatus(status)

	if settledAt.Valid {
		outcome.Status = domain.OutcomeStatus(status)
	}

	return &outcome, nil
}

// GetByMarket retrieves market outcomes by market ID
func (r *MarketOutcomeRepository) GetByMarket(ctx context.Context, marketID string) ([]*domain.Outcome, error) {
	query := `
		SELECT id, outcome_id, market_id, outcome_name, odds, price, status,
			   settled_at, settlement_factor, created_at, updated_at
		FROM market_outcomes
		WHERE market_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outcomes []*domain.Outcome
	for rows.Next() {
		var outcome domain.Outcome
		var name, status string
		var settledAt sql.NullTime
		var settlementFactor float64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&outcome.ID, &outcome.MarketID, &name, &outcome.Odds, &outcome.Price,
			&status, &settledAt, &settlementFactor, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		outcome.Name = name
		outcome.Status = domain.OutcomeStatus(status)

		if settledAt.Valid {
			outcome.Status = domain.OutcomeStatus(status)
		}

		outcomes = append(outcomes, &outcome)
	}

	return outcomes, nil
}

// Update updates a market outcome
func (r *MarketOutcomeRepository) Update(ctx context.Context, outcome *domain.Outcome) error {
	query := `
		UPDATE market_outcomes
		SET outcome_name = $2, odds = $3, price = $4, status = $5,
			settled_at = $6, updated_at = $7
		WHERE id = $1
	`

	var settledAt sql.NullTime
	if outcome.Status == domain.OutcomeStatusWon || outcome.Status == domain.OutcomeStatusLost {
		settledAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		outcome.ID, outcome.Name, outcome.Odds, outcome.Price, string(outcome.Status),
		settledAt, time.Now(),
	)

	return err
}

// UpdateStatus updates market outcome status
func (r *MarketOutcomeRepository) UpdateStatus(ctx context.Context, id string, status domain.OutcomeStatus) error {
	query := `
		UPDATE market_outcomes
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, string(status), time.Now())
	return err
}

// UpdateOdds updates market outcome odds
func (r *MarketOutcomeRepository) UpdateOdds(ctx context.Context, id string, odds decimal.Decimal) error {
	query := `
		UPDATE market_outcomes
		SET odds = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, odds, time.Now())
	return err
}

// Delete deletes a market outcome
func (r *MarketOutcomeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM market_outcomes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetActiveOutcomes retrieves all active market outcomes
func (r *MarketOutcomeRepository) GetActiveOutcomes(ctx context.Context) ([]*domain.Outcome, error) {
	query := `
		SELECT id, outcome_id, market_id, outcome_name, odds, price, status,
			   settled_at, settlement_factor, created_at, updated_at
		FROM market_outcomes
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outcomes []*domain.Outcome
	for rows.Next() {
		var outcome domain.Outcome
		var name, status string
		var settledAt sql.NullTime
		var settlementFactor float64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&outcome.ID, &outcome.MarketID, &name, &outcome.Odds, &outcome.Price,
			&status, &settledAt, &settlementFactor, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		outcome.Name = name
		outcome.Status = domain.OutcomeStatus(status)

		if settledAt.Valid {
			outcome.Status = domain.OutcomeStatus(status)
		}

		outcomes = append(outcomes, &outcome)
	}

	return outcomes, nil
}
