// Package integration provides integration tests for wallet and betting services
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase/tax"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
)

// WalletIntegrationTestSuite provides integration tests for wallet service
type WalletIntegrationTestSuite struct {
	suite.Suite

	ctx           context.Context
	db            *sql.DB
	container     *postgrescontainers.PostgresContainer
	walletService *wallet.Service
	walletRepo    *postgres.WalletRepository
	userRepo      *postgres.UserRepository

	testUser   *domain.User
	testWallet *domain.Wallet
}

// SetupSuite initializes the test suite with PostgreSQL container
func (s *WalletIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Start PostgreSQL container
	container, err := postgrescontainers.RunContainer(s.ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err, "Failed to start PostgreSQL container")
	s.container = container

	// Get database connection
	dbPort, err := s.container.MappedPort(s.ctx, "5432")
	s.Require().NoError(err, "Failed to get mapped port")

	dbHost, err := s.container.Host(s.ctx)
	s.Require().NoError(err, "Failed to get container host")

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		dbHost, dbPort.Port())

	s.db, err = sql.Open("postgres", dsn)
	s.Require().NoError(err, "Failed to open database connection")

	// Test connection
	err = s.db.PingContext(s.ctx)
	s.Require().NoError(err, "Failed to ping database")

	// Run migrations
	s.runMigrations()

	// Initialize repositories
	s.walletRepo = postgres.NewWalletRepository(s.db)
	s.userRepo = postgres.NewUserRepository(s.db)

	// Initialize services
	s.walletService = wallet.New(s.db)
}

// TearDownSuite cleans up the test suite
func (s *WalletIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		if err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}
}

// SetupTest creates test data for each test
func (s *WalletIntegrationTestSuite) SetupTest() {
	// Create test user
	s.testUser = &domain.User{
		ID:           uuid.New(),
		PhoneNumber:  "+254712345678",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CountryCode:  "KE",
		Currency:     "KES",
		FullName:     "Test User",
		DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := s.userRepo.Create(s.ctx, s.testUser)
	s.Require().NoError(err, "Failed to create test user")

	// Create test wallet
	s.testWallet = &domain.Wallet{
		ID:        uuid.New(),
		UserID:    s.testUser.ID,
		Currency:  "KES",
		Balance:   decimal.NewFromFloat(1000.00),
		Version:   1,
		UpdatedAt: time.Now(),
	}
	err = s.walletRepo.Create(s.ctx, s.testWallet)
	s.Require().NoError(err, "Failed to create test wallet")
}

// TearDownTest cleans up test data after each test
func (s *WalletIntegrationTestSuite) TearDownTest() {
	// Clean up in reverse order of dependencies
	_, err := s.db.ExecContext(s.ctx, "DELETE FROM transactions")
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, "DELETE FROM wallets")
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, "DELETE FROM users")
	s.Require().NoError(err)
}

// runMigrations executes database migrations
func (s *WalletIntegrationTestSuite) runMigrations() {
	migrationsDir := "../../migrations"

	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	s.Require().NoError(err, "Failed to read migrations directory")

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		path := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(path)
		s.Require().NoError(err, "Failed to read migration file: %s", file.Name())

		_, err = s.db.ExecContext(s.ctx, string(content))
		s.Require().NoError(err, "Failed to execute migration: %s", file.Name())
	}
}

// TestWalletCredit tests wallet credit functionality
func (s *WalletIntegrationTestSuite) TestWalletCredit() {
	// Arrange
	creditAmount := decimal.NewFromFloat(100.00)
	movement := wallet.Movement{
		UserID:        s.testUser.ID,
		Amount:        creditAmount,
		Type:          domain.TransactionTypeDeposit,
		Description:   "Test deposit",
		ProviderName:  "test",
		ProviderTxnID: uuid.New().String(),
		CountryCode:   "KE",
	}

	// Act
	tx, err := s.walletService.Credit(s.ctx, s.testUser.ID, creditAmount, movement)

	// Assert
	s.Require().NoError(err, "Credit should succeed")
	s.NotNil(tx, "Transaction should be created")
	s.Equal(creditAmount, tx.Amount, "Transaction amount should match")
	s.Equal(domain.TransactionTypeDeposit, tx.Type, "Transaction type should be deposit")
	s.Equal(domain.TransactionStatusCompleted, tx.Status, "Transaction should be completed")

	// Verify wallet balance
	updatedWallet, err := s.walletRepo.GetByUserID(s.ctx, s.testUser.ID)
	s.Require().NoError(err, "Should be able to get updated wallet")
	expectedBalance := decimal.NewFromFloat(1100.00) // 1000 + 100
	s.True(expectedBalance.Equal(updatedWallet.Balance),
		"Expected balance %s, got %s", expectedBalance.String(), updatedWallet.Balance.String())
}

// TestWalletDebit tests wallet debit functionality
func (s *WalletIntegrationTestSuite) TestWalletDebit() {
	// Arrange
	debitAmount := decimal.NewFromFloat(50.00)
	movement := wallet.Movement{
		UserID:        s.testUser.ID,
		Amount:        debitAmount,
		Type:          domain.TransactionTypeWithdrawal,
		Description:   "Test withdrawal",
		ProviderName:  "test",
		ProviderTxnID: uuid.New().String(),
		CountryCode:   "KE",
	}

	// Act
	tx, err := s.walletService.Debit(s.ctx, s.testUser.ID, debitAmount, movement)

	// Assert
	s.Require().NoError(err, "Debit should succeed")
	s.NotNil(tx, "Transaction should be created")
	s.Equal(debitAmount.Neg(), tx.Amount, "Transaction amount should be negative")
	s.Equal(domain.TransactionTypeWithdrawal, tx.Type, "Transaction type should be withdrawal")
	s.Equal(domain.TransactionStatusCompleted, tx.Status, "Transaction should be completed")

	// Verify wallet balance
	updatedWallet, err := s.walletRepo.GetByUserID(s.ctx, s.testUser.ID)
	s.Require().NoError(err, "Should be able to get updated wallet")
	expectedBalance := decimal.NewFromFloat(950.00) // 1000 - 50
	s.True(expectedBalance.Equal(updatedWallet.Balance),
		"Expected balance %s, got %s", expectedBalance.String(), updatedWallet.Balance.String())
}

// TestWalletInsufficientFunds tests debit with insufficient funds
func (s *WalletIntegrationTestSuite) TestWalletInsufficientFunds() {
	// Arrange
	debitAmount := decimal.NewFromFloat(2000.00) // More than balance
	movement := wallet.Movement{
		UserID:        s.testUser.ID,
		Amount:        debitAmount,
		Type:          domain.TransactionTypeWithdrawal,
		Description:   "Test withdrawal",
		ProviderName:  "test",
		ProviderTxnID: uuid.New().String(),
		CountryCode:   "KE",
	}

	// Act
	tx, err := s.walletService.Debit(s.ctx, s.testUser.ID, debitAmount, movement)

	// Assert
	s.Error(err, "Debit should fail with insufficient funds")
	s.Nil(tx, "Transaction should not be created")
	s.Contains(err.Error(), "insufficient funds", "Error should mention insufficient funds")

	// Verify wallet balance is unchanged
	updatedWallet, err := s.walletRepo.GetByUserID(s.ctx, s.testUser.ID)
	s.Require().NoError(err, "Should be able to get wallet")
	expectedBalance := decimal.NewFromFloat(1000.00) // Unchanged
	s.True(expectedBalance.Equal(updatedWallet.Balance),
		"Expected balance %s, got %s", expectedBalance.String(), updatedWallet.Balance.String())
}

// TestConcurrentOperations tests concurrent wallet operations
func (s *WalletIntegrationTestSuite) TestConcurrentOperations() {
	// Arrange
	creditAmount := decimal.NewFromFloat(10.00)
	concurrentOps := 10

	// Act - Run concurrent credits
	done := make(chan bool, concurrentOps)
	for i := range concurrentOps {
		go func() {
			movement := wallet.Movement{
				UserID:        s.testUser.ID,
				Amount:        creditAmount,
				Type:          domain.TransactionTypeDeposit,
				Description:   fmt.Sprintf("Concurrent deposit %d", i),
				ProviderName:  "test",
				ProviderTxnID: uuid.New().String(),
				CountryCode:   "KE",
			}

			_, err := s.walletService.Credit(s.ctx, s.testUser.ID, creditAmount, movement)
			s.Require().NoError(err, "Concurrent credit should succeed")
			done <- true
		}()
	}

	// Wait for all operations to complete
	for range concurrentOps {
		<-done
	}

	// Assert
	updatedWallet, err := s.walletRepo.GetByUserID(s.ctx, s.testUser.ID)
	s.Require().NoError(err, "Should be able to get updated wallet")
	expectedBalance := decimal.NewFromFloat(1100.00) // 1000 + (10 * 10)
	s.True(expectedBalance.Equal(updatedWallet.Balance),
		"Expected balance %s, got %s", expectedBalance.String(), updatedWallet.Balance.String())
}

// TestWalletService runs the wallet integration test suite
func TestWalletService(t *testing.T) {
	suite.Run(t, new(WalletIntegrationTestSuite))
}

// TestWalletIntegrationWithTax tests wallet integration with tax engine
func (s *WalletIntegrationTestSuite) TestWalletIntegrationWithTax() {
	// Arrange
	taxEngine := tax.Default()
	grossPayout := decimal.NewFromFloat(200.00)
	stake := decimal.NewFromFloat(100.00)

	// Act - Apply tax
	payoutBreakdown := taxEngine.ApplyPayoutTax("KE", grossPayout, stake)

	// Assert
	s.NotNil(payoutBreakdown, "Payout breakdown should not be nil")
	s.True(payoutBreakdown.NetPayout.LessThan(grossPayout), "Net payout should be less than gross payout")
	s.True(payoutBreakdown.WinningsTax.GreaterThan(decimal.Zero), "Winnings tax should be positive")

	// Verify tax calculation (20% of winnings)
	expectedWinnings := grossPayout.Sub(stake)                        // 100
	expectedTax := expectedWinnings.Mul(decimal.NewFromFloat(0.20))   // 20
	expectedNetPayout := stake.Add(expectedWinnings.Sub(expectedTax)) // 100 + 80 = 180

	s.True(expectedTax.Equal(payoutBreakdown.WinningsTax),
		"Expected tax %s, got %s", expectedTax.String(), payoutBreakdown.WinningsTax.String())
	s.True(expectedNetPayout.Equal(payoutBreakdown.NetPayout),
		"Expected net payout %s, got %s", expectedNetPayout.String(), payoutBreakdown.NetPayout.String())
}
