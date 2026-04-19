package ke

import (
	"errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrMPesaTimeout       = errors.New("mpesa request timeout")
	ErrInsufficientFunds  = errors.New("mpesa insufficient funds")
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
)

// MPesaConfig holds Safaricom Daraja API credentials
type MPesaConfig struct {
	ConsumerKey        string
	ConsumerSecret     string
	ShortCode          string // Paybill/Till number
	PassKey            string // For STK Push
	InitiatorName      string // For B2C
	SecurityCredential string // Encrypted initiator password
	Environment        string // sandbox or production
}

// MPesaClient handles all M-Pesa operations
type MPesaClient struct {
	config     MPesaConfig
	httpClient *http.Client
	baseURL    string
}

// STKPushRequest represents an STK Push request
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallbackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
	Remark            string `json:"Remark"`
}

// STKPushResponse represents STK Push response
type STKPushResponse struct {
	MerchantRequestID string `json:"MerchantRequestID"`
	CheckoutRequestID string `json:"CheckoutRequestID"`
	ResponseCode      string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage   string `json:"CustomerMessage"`
}

// B2CRequest represents a B2C payment request
type B2CRequest struct {
	InitiatorName      string `json:"InitiatorName"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	Amount             string `json:"Amount"`
	PartyA             string `json:"PartyA"`
	PartyB             string `json:"PartyB"`
	Remarks            string `json:"Remarks"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	ResultURL          string `json:"ResultURL"`
	Occasion           string `json:"Occasion"`
}

// B2CResponse represents B2C payment response
type B2CResponse struct {
	ConversationID string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ResponseDescription string `json:"ResponseDescription"`
	ResponseCode string `json:"ResponseCode"`
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	PhoneNumber string          `json:"phone_number"`
	Amount     decimal.Decimal `json:"amount"`
	Reference  string          `json:"reference"`
	Description string          `json:"description"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	TransactionID string        `json:"transaction_id,omitempty"`
	Amount     decimal.Decimal `json:"amount,omitempty"`
	Status     string          `json:"status,omitempty"`
}

// TransactionStatusRequest represents a transaction status request
type TransactionStatusRequest struct {
	TransactionID string `json:"transaction_id"`
}

// TransactionStatusResponse represents transaction status response
type TransactionStatusResponse struct {
	Success         bool            `json:"success"`
	Message         string          `json:"message"`
	TransactionID   string          `json:"transaction_id,omitempty"`
	Status          string          `json:"status,omitempty"`
	Amount          decimal.Decimal `json:"amount,omitempty"`
	Currency        string          `json:"currency,omitempty"`
	CreatedAt       time.Time       `json:"created_at,omitempty"`
	CompletedAt     time.Time       `json:"completed_at,omitempty"`
}

// AccountBalanceRequest represents an account balance request
type AccountBalanceRequest struct {
	InitiatorName      string `json:"InitiatorName"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	PartyA             string `json:"PartyA"`
	IdentifierType     string `json:"IdentifierType"`
	Remarks            string `json:"Remarks"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	ResultURL          string `json:"ResultURL"`
}

// AccountBalanceResponse represents account balance response
type AccountBalanceResponse struct {
	Success         bool            `json:"success"`
	Message         string          `json:"message"`
	Balance         decimal.Decimal `json:"balance,omitempty"`
	Currency        string          `json:"currency,omitempty"`
	LastUpdated     time.Time       `json:"last_updated,omitempty"`
}

// MPesaMetrics represents client metrics
type MPesaMetrics struct {
	TotalTransactions    int64           `json:"total_transactions"`
	SuccessfulTransactions int64        `json:"successful_transactions"`
	FailedTransactions   int64           `json:"failed_transactions"`
	TotalAmount          decimal.Decimal `json:"total_amount"`
	SuccessRate          decimal.Decimal `json:"success_rate"`
	AverageAmount        decimal.Decimal `json:"average_amount"`
	LastTransactionTime  time.Time       `json:"last_transaction_time"`
}

// WebhookEvent represents an M-Pesa webhook event
type WebhookEvent struct {
	TransactionType string      `json:"TransactionType"`
	TransID         string      `json:"TransID"`
	TransTime       string      `json:"TransTime"`
	Amount          string      `json:"Amount"`
	BusinessShortCode string    `json:"BusinessShortCode"`
	BillRefNumber   string      `json:"BillRefNumber"`
	InvoiceNumber   string      `json:"InvoiceNumber"`
	OrgAccountBalance string    `json:"OrgAccountBalance"`
	ThirdPartyTransID string     `json:"ThirdPartyTransID"`
	MSISDN          string      `json:"MSISDN"`
	FirstName       string      `json:"FirstName"`
	MiddleName      string      `json:"MiddleName"`
	LastName        string      `json:"LastName"`
}

// PaymentResult represents the result of a payment operation
type PaymentResult struct {
	TransactionID string          `json:"transaction_id"`
	Status        string          `json:"status"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	PhoneNumber   string          `json:"phone_number"`
	Reference     string          `json:"reference"`
	Description   string          `json:"description"`
	CreatedAt     time.Time       `json:"created_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage  string          `json:"error_message,omitempty"`
}
