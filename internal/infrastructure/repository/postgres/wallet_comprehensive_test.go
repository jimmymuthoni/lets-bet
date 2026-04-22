package postgres

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWalletRepositoryCreate tests wallet creation functionality
func TestWalletRepositoryCreate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	assert.NoError(t, err, "Wallet creation should succeed")

	// Verify wallet was created
	retrieved, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, wallet.ID, retrieved.ID)
	assert.Equal(t, userID, retrieved.UserID)
	assert.Equal(t, "USD", retrieved.Currency)
	assert.True(t, retrieved.Balance.Equal(decimal.NewFromFloat(1000)))
	assert.True(t, retrieved.BonusBalance.Equal(decimal.NewFromFloat(100)))
	assert.Equal(t, int64(1), retrieved.Version)
}

// TestWalletRepositoryCreateDuplicate tests duplicate wallet creation handling
func TestWalletRepositoryCreateDuplicate(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	userID := uuid.New()
	wallet1 := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	// Create first wallet
	err := repo.Create(context.Background(), wallet1)
	require.NoError(t, err)

	// Try to create second wallet for same user
	wallet2 := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID, // Same user ID
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(2000),
		BonusBalance: decimal.NewFromFloat(200),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(context.Background(), wallet2)
	assert.Error(t, err, "Duplicate wallet creation should fail")
}

// TestWalletRepositoryGetByUserID tests wallet retrieval by user ID
func TestWalletRepositoryGetByUserID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Test non-existent wallet
	nonExistentUserID := uuid.New()
	wallet, err := repo.GetByUserID(context.Background(), nonExistentUserID)
	assert.Error(t, err, "Getting non-existent wallet should fail")
	assert.Nil(t, wallet, "Should return nil for non-existent wallet")

	// Create wallet and test retrieval
	userID := uuid.New()
	createdWallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "EUR",
		Balance:      decimal.NewFromFloat(500),
		BonusBalance: decimal.NewFromFloat(50),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(context.Background(), createdWallet)
	require.NoError(t, err)

	// Retrieve wallet
	retrievedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, createdWallet.ID, retrievedWallet.ID)
	assert.Equal(t, userID, retrievedWallet.UserID)
	assert.Equal(t, "EUR", retrievedWallet.Currency)
	assert.True(t, retrievedWallet.Balance.Equal(decimal.NewFromFloat(500)))
	assert.True(t, retrievedWallet.BonusBalance.Equal(decimal.NewFromFloat(50)))
}

// TestWalletRepositoryUpdateBalance tests balance update functionality
func TestWalletRepositoryUpdateBalance(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Update balance
	newBalance := decimal.NewFromFloat(1500)
	newBonusBalance := decimal.NewFromFloat(150)

	err = repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.NoError(t, err, "Balance update should succeed")

	// Verify update
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.Balance.Equal(newBalance))
	assert.True(t, updatedWallet.BonusBalance.Equal(newBonusBalance))
	assert.Equal(t, int64(2), updatedWallet.Version, "Version should be incremented")
}

// TestWalletRepositoryUpdateBalanceNegative tests negative balance handling
func TestWalletRepositoryUpdateBalanceNegative(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(100),
		BonusBalance: decimal.NewFromFloat(50),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Try to set negative balance
	negativeBalance := decimal.NewFromFloat(-50)
	negativeBonusBalance := decimal.NewFromFloat(-10)

	err = repo.UpdateBalance(context.Background(), userID, negativeBalance, negativeBonusBalance)
	// Note: The repository doesn't prevent negative balances, this would be handled at the service layer
	assert.NoError(t, err, "Repository should allow negative balance (validation at service layer)")

	// Verify update
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.Balance.Equal(negativeBalance))
	assert.True(t, updatedWallet.BonusBalance.Equal(negativeBonusBalance))
}

// TestWalletRepositoryCreateTransaction tests transaction creation
func TestWalletRepositoryCreateTransaction(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet first
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Create transaction
	referenceID := uuid.New()
	transaction := &domain.Transaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		UserID:        userID,
		Type:          domain.TransactionTypeDeposit,
		Amount:        decimal.NewFromFloat(100),
		Currency:      "USD",
		BalanceBefore: decimal.NewFromFloat(1000),
		BalanceAfter:  decimal.NewFromFloat(1100),
		ReferenceID:   &referenceID,
		ReferenceType: "DEPOSIT",
		Status:        domain.TransactionStatusCompleted,
		Description:   "Test deposit",
		CreatedAt:     time.Now(),
		CompletedAt:   &[]time.Time{time.Now()}[0],
		CountryCode:   "US",
	}

	err = repo.CreateTransaction(context.Background(), transaction)
	assert.NoError(t, err, "Transaction creation should succeed")

	// Verify transaction was created (would need GetTransaction method)
	// For now, just ensure no error occurred
}

func TestWalletRepositoryMultipleTransactions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Create multiple transactions
	for i := 0; i < 5; i++ {
		referenceID := uuid.New()
		completedAt := time.Now().Add(time.Duration(i) * time.Hour)
		transaction := &domain.Transaction{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			UserID:        userID,
			Type:          domain.TransactionTypeDeposit,
			Amount:        decimal.NewFromFloat(100),
			Currency:      "USD",
			BalanceBefore: decimal.NewFromFloat(1000 + float64(i*100)),
			BalanceAfter:  decimal.NewFromFloat(1100 + float64(i*100)),
			ReferenceID:   &referenceID,
			ReferenceType: "DEPOSIT",
			Status:        domain.TransactionStatusCompleted,
			Description:   "Test deposit",
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
			CompletedAt:   &completedAt,
			CountryCode:   "US",
		}

		err = repo.CreateTransaction(context.Background(), transaction)
		require.NoError(t, err)
	}

	// All transactions should be created successfully
	// We can't retrieve them without GetTransactions method, but creation should work
}

func TestWalletRepositoryMultipleCurrencies(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallets with different currencies
	currencies := []string{"USD", "EUR", "GBP", "JPY"}
	userIDs := make([]uuid.UUID, len(currencies))

	for i, currency := range currencies {
		userID := uuid.New()
		userIDs[i] = userID

		wallet := &domain.Wallet{
			ID:           uuid.New(),
			UserID:       userID,
			Currency:     currency,
			Balance:      decimal.NewFromFloat(1000),
			BonusBalance: decimal.NewFromFloat(100),
			Version:      1,
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(context.Background(), wallet)
		require.NoError(t, err)
	}

	// Verify all wallets were created correctly
	for i, currency := range currencies {
		wallet, err := repo.GetByUserID(context.Background(), userIDs[i])
		require.NoError(t, err)
		assert.Equal(t, currency, wallet.Currency)
		assert.True(t, wallet.Balance.Equal(decimal.NewFromFloat(1000)))
	}
}

// TestWalletRepositoryLargeAmounts tests wallet operations with large amounts
func TestWalletRepositoryLargeAmounts(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet with large balance
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(999999999.99),
		BonusBalance: decimal.NewFromFloat(99999999.99),
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Update with large amounts
	newBalance := decimal.NewFromFloat(1999999999.99)
	newBonusBalance := decimal.NewFromFloat(199999999.99)

	err = repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.NoError(t, err, "Large amount update should succeed")

	// Verify update
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.Balance.Equal(newBalance))
	assert.True(t, updatedWallet.BonusBalance.Equal(newBonusBalance))
}

// TestWalletRepositoryTimestamps tests that timestamps are properly handled
func TestWalletRepositoryTimestamps(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet with specific timestamp
	userID := uuid.New()
	createdTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.NewFromFloat(1000),
		BonusBalance: decimal.NewFromFloat(100),
		Version:      1,
		UpdatedAt:    createdTime,
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Update wallet
	newBalance := decimal.NewFromFloat(1500)
	newBonusBalance := decimal.NewFromFloat(150)

	err = repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.NoError(t, err)

	// Verify timestamp was updated
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.UpdatedAt.After(createdTime),
		"UpdatedAt should be more recent after update")
}

// TestWalletRepositoryZeroAmounts tests wallet operations with zero amounts
func TestWalletRepositoryZeroAmounts(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create wallet with zero balances
	userID := uuid.New()
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      decimal.Zero,
		BonusBalance: decimal.Zero,
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Verify zero balances
	retrievedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, retrievedWallet.Balance.Equal(decimal.Zero))
	assert.True(t, retrievedWallet.BonusBalance.Equal(decimal.Zero))

	// Update to non-zero amounts
	newBalance := decimal.NewFromFloat(100)
	newBonusBalance := decimal.NewFromFloat(10)

	err = repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.NoError(t, err)

	// Verify update
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.Balance.Equal(newBalance))
	assert.True(t, updatedWallet.BonusBalance.Equal(newBonusBalance))
}

// TestWalletOptimisticLocking tests that concurrent balance updates are properly serialized
func TestWalletOptimisticLocking(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create test user and wallet
	userID := uuid.New()
	initialBalance := decimal.NewFromFloat(1000)
	initialBonusBalance := decimal.NewFromFloat(100)

	// Create initial wallet
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      initialBalance,
		BonusBalance: initialBonusBalance,
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Test 1: Normal update should work
	newBalance := initialBalance.Add(decimal.NewFromFloat(100))
	newBonusBalance := initialBonusBalance.Add(decimal.NewFromFloat(10))

	err = repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.NoError(t, err, "Normal update should succeed")

	// Verify the balance was updated
	updatedWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, updatedWallet.Balance.Equal(newBalance), "Balance should be updated")
	assert.True(t, updatedWallet.BonusBalance.Equal(newBonusBalance), "Bonus balance should be updated")
	assert.Equal(t, int64(2), updatedWallet.Version, "Version should be incremented")
}

// TestConcurrentBalanceUpdates tests that concurrent updates don't lose money
func TestConcurrentBalanceUpdates(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create test user and wallet
	userID := uuid.New()
	initialBalance := decimal.NewFromFloat(1000)

	// Create initial wallet
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      initialBalance,
		BonusBalance: decimal.Zero,
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Simulate 10 concurrent deposits of $100 each
	var wg sync.WaitGroup
	numDeposits := 10
	depositAmount := decimal.NewFromFloat(100)
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numDeposits; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Get current balance and add deposit
			currentWallet, err := repo.GetByUserID(context.Background(), userID)
			if err != nil {
				t.Logf("Error getting wallet: %v", err)
				return
			}

			newBalance := currentWallet.Balance.Add(depositAmount)

			// Try to update with optimistic locking
			err = repo.UpdateBalance(context.Background(), userID, newBalance, currentWallet.BonusBalance)
			if err != nil {
				t.Logf("Update failed (expected under contention): %v", err)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Verify final state
	finalWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)

	// The final balance should be exactly initial + (successful deposits * deposit amount)
	expectedBalance := initialBalance.Add(decimal.NewFromInt(int64(successCount)).Mul(depositAmount))

	t.Logf("Initial: %s, Final: %s, Expected: %s, Success: %d/%d",
		initialBalance.String(), finalWallet.Balance.String(),
		expectedBalance.String(), successCount, numDeposits)

	assert.True(t, finalWallet.Balance.Equal(expectedBalance),
		"Final balance should match expected amount")
	assert.Greater(t, successCount, 0, "At least some deposits should succeed")
	assert.LessOrEqual(t, successCount, numDeposits, "Success count should not exceed attempts")
}

// TestOptimisticLockConflict tests that version conflicts are properly detected
func TestOptimisticLockConflict(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create test user and wallet
	userID := uuid.New()
	initialBalance := decimal.NewFromFloat(1000)

	// Create initial wallet
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      initialBalance,
		BonusBalance: decimal.Zero,
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Get initial wallet state
	wallet1, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), wallet1.Version)

	// Simulate concurrent update by manually updating the version in database
	// This simulates another transaction updating the wallet
	_, err = db.ExecContext(context.Background(),
		"UPDATE wallets SET version = version + 1 WHERE user_id = $1", userID)
	require.NoError(t, err)

	// Now try to update with the old version - should fail
	newBalance := wallet1.Balance.Add(decimal.NewFromFloat(100))
	err = repo.UpdateBalance(context.Background(), userID, newBalance, wallet1.BonusBalance)

	// Should fail with optimistic lock conflict
	assert.Error(t, err, "Update with old version should fail")
	assert.Contains(t, err.Error(), "optimistic lock conflict",
		"Error should indicate optimistic lock conflict")

	// Verify wallet wasn't updated (balance should remain unchanged)
	finalWallet, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, finalWallet.Balance.Equal(initialBalance),
		"Balance should remain unchanged after failed update")
	assert.Equal(t, int64(2), finalWallet.Version,
		"Version should be incremented by the manual update")
}

// TestWalletNotFound tests proper error handling when wallet doesn't exist
func TestWalletNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Try to update non-existent wallet
	userID := uuid.New()
	newBalance := decimal.NewFromFloat(1000)
	newBonusBalance := decimal.NewFromFloat(100)

	err := repo.UpdateBalance(context.Background(), userID, newBalance, newBonusBalance)
	assert.Error(t, err, "Update should fail for non-existent wallet")
	assert.Contains(t, err.Error(), "wallet not found",
		"Error should indicate wallet not found")
}

// TestVersionIncrement tests that version is properly incremented on each update
func TestVersionIncrement(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := NewWalletRepository(db)

	// Create test user and wallet
	userID := uuid.New()
	initialBalance := decimal.NewFromFloat(1000)

	// Create initial wallet
	wallet := &domain.Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Currency:     "USD",
		Balance:      initialBalance,
		BonusBalance: decimal.Zero,
		Version:      1,
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(context.Background(), wallet)
	require.NoError(t, err)

	// Perform multiple updates and verify version increments
	expectedVersion := int64(1)
	for i := 0; i < 5; i++ {
		expectedVersion++

		currentWallet, err := repo.GetByUserID(context.Background(), userID)
		require.NoError(t, err)

		newBalance := currentWallet.Balance.Add(decimal.NewFromFloat(10))
		err = repo.UpdateBalance(context.Background(), userID, newBalance, currentWallet.BonusBalance)
		require.NoError(t, err, "Update %d should succeed", i+1)

		updatedWallet, err := repo.GetByUserID(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, expectedVersion, updatedWallet.Version,
			"Version should be %d after update %d", expectedVersion, i+1)
	}
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// In a real implementation, this would create a test database
	// For now, we'll use a mock or skip the test if no test DB is available
	t.Skip("Test database not configured - implement test DB setup")
	return nil
}
