// Package gh provides Ghana-specific payment adapters
package gh

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// FlutterwaveConfig provides configuration for Flutterwave Ghana adapter
type FlutterwaveConfig struct {
	SecretKey     string `json:"secret_key"`
	PublicKey     string `json:"public_key"`
	Environment   string `json:"environment"` // "live" or "test"
	BaseURL       string `json:"base_url"`
	WebhookSecret string `json:"webhook_secret"`
	Currency      string `json:"currency"` // GHS for Ghana
}

// DefaultFlutterwaveConfig returns default configuration for Ghana
func DefaultFlutterwaveConfig() *FlutterwaveConfig {
	return &FlutterwaveConfig{
		Environment: "test",
		BaseURL:     "https://api.flutterwave.com/v3",
		Currency:    "GHS",
	}
}

// FlutterwaveAdapter provides Flutterwave payment processing for Ghana
type FlutterwaveAdapter struct {
	config *FlutterwaveConfig
	client *http.Client
}

// PaymentRequest represents a Flutterwave payment request
type PaymentRequest struct {
	TxRef         string          `json:"tx_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	CustomerEmail string          `json:"customer_email"`
	CustomerPhone string          `json:"customer_phone"`
	CustomerName  string          `json:"customer_name"`
	RedirectURL   string          `json:"redirect_url"`
}

// TransactionVerificationResponse represents a verification response
type TransactionVerificationResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    TransactionVerificationData `json:"data"`
}

// TransactionVerificationData contains verification data
type TransactionVerificationData struct {
	ID              string          `json:"id"`
	TxRef           string          `json:"tx_ref"`
	FlwRef          string          `json:"flw_ref"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	PaymentType     string          `json:"payment_type"`
	Customer        CustomerInfo    `json:"customer"`
	CreatedAt       time.Time       `json:"created_at"`
	ChargedAmount   decimal.Decimal `json:"charged_amount"`
	AppFee          decimal.Decimal `json:"app_fee"`
	MerchantFee     decimal.Decimal `json:"merchant_fee"`
	ProcessResponse ProcessResponse `json:"process_response"`
}

// ProcessResponse contains processing response details
type ProcessResponse struct {
	Response string `json:"response"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	ID     string          `json:"id"`
	Amount decimal.Decimal `json:"amount"`
	Reason string          `json:"reason"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    RefundData `json:"data"`
}

// RefundData contains refund response data
type RefundData struct {
	ID              string          `json:"id"`
	TxRef           string          `json:"tx_ref"`
	FlwRef          string          `json:"flw_ref"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	RefundRef       string          `json:"refund_ref"`
	CreatedAt       time.Time       `json:"created_at"`
	ProcessResponse ProcessResponse `json:"process_response"`
}

// WebhookEvent represents a Flutterwave webhook event
type WebhookEvent struct {
	Event string      `json:"event"`
	Data  WebhookData `json:"data"`
}

// WebhookData contains webhook event data
type WebhookData struct {
	ID            string          `json:"id"`
	TxRef         string          `json:"tx_ref"`
	FlwRef        string          `json:"flw_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	PaymentType   string          `json:"payment_type"`
	Customer      CustomerInfo    `json:"customer"`
	CreatedAt     time.Time       `json:"created_at"`
	ChargedAmount decimal.Decimal `json:"charged_amount"`
	AppFee        decimal.Decimal `json:"app_fee"`
	MerchantFee   decimal.Decimal `json:"merchant_fee"`
}

// FlutterwaveError represents a Flutterwave API error
type FlutterwaveError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// FlutterwaveMetrics represents Flutterwave adapter metrics
type FlutterwaveMetrics struct {
	TotalTransactions      int64           `json:"total_transactions"`
	SuccessfulTransactions int64           `json:"successful_transactions"`
	FailedTransactions     int64           `json:"failed_transactions"`
	TotalAmount            decimal.Decimal `json:"total_amount"`
	SuccessRate            decimal.Decimal `json:"success_rate"`
	AverageAmount          decimal.Decimal `json:"average_amount"`
	LastTransactionTime    time.Time       `json:"last_transaction_time"`
}
