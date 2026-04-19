package ke

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// AccessTokenResponse represents the response from M-Pesa OAuth API
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewMPesaClient creates a new M-Pesa client
func NewMPesaClient(config MPesaConfig) *MPesaClient {
	baseURL := "https://api.safaricom.co.ke"
	if config.Environment == "sandbox" {
		baseURL = "https://sandbox.safaricom.co.ke"
	}

	return &MPesaClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

// GetAccessToken retrieves OAuth token from M-Pesa API
func (c *MPesaClient) GetAccessToken(ctx context.Context) (string, error) {
	url := c.baseURL + "/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.config.ConsumerKey, c.config.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %d", resp.StatusCode)
	}

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// InitiateDeposit triggers STK Push (Lipa Na M-Pesa Online)
func (c *MPesaClient) InitiateDeposit(ctx context.Context, phoneNumber string, amount decimal.Decimal, reference string) (*STKPushResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Generate password: Base64(ShortCode + PassKey + Timestamp)
	passwordStr := c.config.ShortCode + c.config.PassKey + timestamp
	password := base64.StdEncoding.EncodeToString([]byte(passwordStr))

	// Format phone number (254XXXXXXXXX)
	if len(phoneNumber) == 10 && phoneNumber[0] == '0' {
		phoneNumber = "254" + phoneNumber[1:]
	}

	// Create STK push request with the correct structure
	reqBody := map[string]any{
		"BusinessShortCode": c.config.ShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            amount.StringFixed(0),
		"PartyA":            phoneNumber,
		"PartyB":            c.config.ShortCode,
		"PhoneNumber":       phoneNumber,
		"CallBackURL":       "https://yourdomain.com/api/mpesa/callback",
		"AccountReference":  reference,
		"TransactionDesc":   "Deposit",
	}

	jsonData, _ := json.Marshal(reqBody)
	url := c.baseURL + "/mpesa/stkpush/v1/processrequest"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("STK push failed: %d - %s", resp.StatusCode, string(body))
	}

	var stkResp STKPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&stkResp); err != nil {
		return nil, err
	}

	return &stkResp, nil
}

// ProcessWithdrawal processes B2C payment (withdrawal)
func (c *MPesaClient) ProcessWithdrawal(ctx context.Context, phoneNumber string, amount decimal.Decimal, reference string) (*B2CResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Format phone number (254XXXXXXXXX)
	if len(phoneNumber) == 10 && phoneNumber[0] == '0' {
		phoneNumber = "254" + phoneNumber[1:]
	}

	reqBody := map[string]any{
		"InitiatorName":      c.config.InitiatorName,
		"SecurityCredential": c.config.SecurityCredential,
		"CommandID":          "BusinessPayment",
		"Amount":             amount.StringFixed(0),
		"PartyA":             c.config.ShortCode,
		"PartyB":             phoneNumber,
		"Remarks":            reference,
		"QueueTimeOutURL":    "https://yourdomain.com/api/mpesa/timeout",
		"ResultURL":          "https://yourdomain.com/api/mpesa/result",
		"Occasion":           "Withdrawal",
	}

	jsonData, _ := json.Marshal(reqBody)
	url := c.baseURL + "/mpesa/b2c/v1/paymentrequest"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("B2C payment failed: %d - %s", resp.StatusCode, string(body))
	}

	var b2cResp B2CResponse
	if err := json.NewDecoder(resp.Body).Decode(&b2cResp); err != nil {
		return nil, err
	}

	return &b2cResp, nil
}

// CheckTransactionStatus checks the status of a transaction
func (c *MPesaClient) CheckTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatusResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"Initiator":          c.config.InitiatorName,
		"SecurityCredential": c.config.SecurityCredential,
		"CommandID":          "TransactionStatusQuery",
		"TransactionID":      transactionID,
		"PartyA":             c.config.ShortCode,
		"IdentifierType":     "4",
		"ResultURL":          "https://yourdomain.com/api/mpesa/result",
		"QueueTimeOutURL":    "https://yourdomain.com/api/mpesa/timeout",
		"Remarks":            "Status check",
		"Occasion":           "",
	}

	jsonData, _ := json.Marshal(reqBody)
	url := c.baseURL + "/mpesa/transactionstatus/v1/query"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status check failed: %d - %s", resp.StatusCode, string(body))
	}

	var statusResp TransactionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}

// GetAccountBalance retrieves account balance
func (c *MPesaClient) GetAccountBalance(ctx context.Context) (*AccountBalanceResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]any{
		"Initiator":          c.config.InitiatorName,
		"SecurityCredential": c.config.SecurityCredential,
		"CommandID":          "AccountBalance",
		"PartyA":             c.config.ShortCode,
		"IdentifierType":     "4",
		"Remarks":            "Balance check",
		"QueueTimeOutURL":    "https://yourdomain.com/api/mpesa/timeout",
		"ResultURL":          "https://yourdomain.com/api/mpesa/result",
	}

	jsonData, _ := json.Marshal(reqBody)
	url := c.baseURL + "/mpesa/accountbalance/v1/query"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("balance check failed: %d - %s", resp.StatusCode, string(body))
	}

	var balanceResp AccountBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return nil, err
	}

	return &balanceResp, nil
}

// GetMetrics returns client metrics
func (c *MPesaClient) GetMetrics(ctx context.Context) (*MPesaMetrics, error) {
	// In a real implementation, these would be calculated from actual data
	return &MPesaMetrics{
		TotalTransactions:      1000,
		SuccessfulTransactions: 950,
		FailedTransactions:     50,
		TotalAmount:            decimal.NewFromInt(100000),
		SuccessRate:            decimal.NewFromFloat(0.95),
		AverageAmount:          decimal.NewFromInt(100),
		LastTransactionTime:    time.Now().Add(-1 * time.Hour),
	}, nil
}
