// Package gh provides Ghana-specific payment adapters
package gh

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// NewFlutterwaveAdapter creates a new Flutterwave adapter for Ghana
func NewFlutterwaveAdapter(config *FlutterwaveConfig) *FlutterwaveAdapter {
	if config == nil {
		config = DefaultFlutterwaveConfig()
	}

	return &FlutterwaveAdapter{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePayment creates a new payment with Flutterwave
func (fwa *FlutterwaveAdapter) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Convert to Flutterwave request format
	fwReq := &FlutterwavePaymentRequest{
		TxRef:          req.TxRef,
		Amount:         req.Amount.StringFixed(2),
		Currency:       req.Currency,
		CustomerEmail:  req.CustomerEmail,
		CustomerPhone:  req.CustomerPhone,
		PaymentOptions: "mobilemoneyghana",
		RedirectURL:    req.RedirectURL,
		Customization: Customization{
			Title:       "Betting Platform Payment",
			Description: "Payment for betting services",
		},
	}

	// Make API request
	resp, err := fwa.makeRequest(ctx, "POST", "/payments", fwReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	var fwResp FlutterwavePaymentResponse
	if err := json.Unmarshal(resp, &fwResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to domain response
	paymentResp := &PaymentResponse{
		Status:  fwResp.Status,
		Message: fwResp.Message,
		Data: PaymentData{
			ID:            fmt.Sprintf("%d", fwResp.Data.ID),
			TxRef:         fwResp.Data.TxRef,
			FlwRef:        fwResp.Data.FlwRef,
			Amount:        decimal.RequireFromString(fwResp.Data.Amount),
			Currency:      fwResp.Data.Currency,
			Status:        fwResp.Status,
			PaymentType:   "mobilemoneygh",
			Customer: CustomerInfo{
				ID:    fwResp.Data.ID,
				Name:  fwResp.Data.Customer.Name,
				Email: fwResp.Data.Customer.Email,
				Phone: fwResp.Data.Customer.Phone,
			},
			CreatedAt:     time.Now(),
			ChargedAmount: decimal.RequireFromString(fwResp.Data.Amount),
		},
	}

	return paymentResp, nil
}

// VerifyPayment verifies a payment with Flutterwave
func (fwa *FlutterwaveAdapter) VerifyPayment(ctx context.Context, txRef string) (*TransactionVerificationResponse, error) {
	// Make verification request
	resp, err := fwa.makeRequest(ctx, "GET", "/transactions/"+txRef+"/verify", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to verify payment: %w", err)
	}

	var fwResp FlutterwaveVerifyResponse
	if err := json.Unmarshal(resp, &fwResp); err != nil {
		return nil, fmt.Errorf("failed to parse verification response: %w", err)
	}

	// Convert to domain response
	verificationResp := &TransactionVerificationResponse{
		Status:  fwResp.Status,
		Message: fwResp.Message,
		Data: TransactionVerificationData{
			ID:            fmt.Sprintf("%d", fwResp.Data.ID),
			TxRef:         fwResp.Data.TxRef,
			FlwRef:        fwResp.Data.FlwRef,
			Amount:        decimal.RequireFromString(fwResp.Data.Amount),
			Currency:      fwResp.Data.Currency,
			Status:        fwResp.Data.Status,
			PaymentType:   fwResp.Data.PaymentType,
			Customer: CustomerInfo{
				ID:    fwResp.Data.ID,
				Name:  fwResp.Data.Customer.Name,
				Email: fwResp.Data.Customer.Email,
				Phone: fwResp.Data.Customer.Phone,
			},
			CreatedAt:     time.Now(),
			ChargedAmount: decimal.RequireFromString(fwResp.Data.ChargedAmount),
			AppFee:        decimal.RequireFromString(fwResp.Data.AppFee),
			MerchantFee:   decimal.RequireFromString(fwResp.Data.MerchantFee),
		},
	}

	return verificationResp, nil
}

// ProcessMobileMoneyPayment processes a mobile money payment
func (fwa *FlutterwaveAdapter) ProcessMobileMoneyPayment(ctx context.Context, req *FlutterwaveMobileMoneyRequest) (*FlutterwaveMobileMoneyResponse, error) {
	resp, err := fwa.makeRequest(ctx, "POST", "/charges", req)
	if err != nil {
		return nil, fmt.Errorf("failed to process mobile money payment: %w", err)
	}

	var fwResp FlutterwaveMobileMoneyResponse
	if err := json.Unmarshal(resp, &fwResp); err != nil {
		return nil, fmt.Errorf("failed to parse mobile money response: %w", err)
	}

	return &fwResp, nil
}

// ProcessPayout processes a payout to a bank account
func (fwa *FlutterwaveAdapter) ProcessPayout(ctx context.Context, req *FlutterwavePayoutRequest) (*FlutterwavePayoutResponse, error) {
	resp, err := fwa.makeRequest(ctx, "POST", "/transfers", req)
	if err != nil {
		return nil, fmt.Errorf("failed to process payout: %w", err)
	}

	var fwResp FlutterwavePayoutResponse
	if err := json.Unmarshal(resp, &fwResp); err != nil {
		return nil, fmt.Errorf("failed to parse payout response: %w", err)
	}

	return &fwResp, nil
}

// makeRequest makes an HTTP request to Flutterwave API
func (fwa *FlutterwaveAdapter) makeRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := fwa.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+fwa.config.SecretKey)

	resp, err := fwa.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// verifyWebhookSignature verifies Flutterwave webhook signature
func (fwa *FlutterwaveAdapter) verifyWebhookSignature(payload []byte, signature string) bool {
	if fwa.config.WebhookSecret == "" {
		return true // Skip verification if no secret configured
	}

	h := hmac.New(sha256.New, []byte(fwa.config.WebhookSecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return signature == expectedSignature
}

// GetMetrics returns Flutterwave adapter metrics
func (fwa *FlutterwaveAdapter) GetMetrics(ctx context.Context) (*FlutterwaveMetrics, error) {
	// In a real implementation, these would be calculated from actual data
	return &FlutterwaveMetrics{
		TotalTransactions:       1000,
		SuccessfulTransactions: 950,
		FailedTransactions:     50,
		TotalAmount:            decimal.NewFromInt(50000),
		SuccessRate:            decimal.NewFromFloat(0.95),
		AverageAmount:          decimal.NewFromInt(50),
		LastTransactionTime:    time.Now().Add(-1 * time.Hour),
	}, nil
}
