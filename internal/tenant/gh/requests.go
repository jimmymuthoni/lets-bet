// Package gh provides Ghana-specific payment adapters
package gh

type FlutterwavePaymentRequest struct {
	TxRef          string         `json:"tx_ref"`
	Amount         string         `json:"amount"`
	Currency       string         `json:"currency"`
	CustomerEmail  string         `json:"customer_email"`
	CustomerPhone  string         `json:"customer_phone,omitempty"`
	PaymentOptions string         `json:"payment_options"` // "card, banktransfer, mobilemoneyghana"
	RedirectURL    string         `json:"redirect_url"`
	PaymentPlan    int            `json:"payment_plan,omitempty"`
	SubAccounts    []SubAccount   `json:"subaccounts,omitempty"`
	Meta           map[string]any `json:"meta,omitempty"`
	Customization  Customization  `json:"customization"`
}

// SubAccount represents a Flutterwave subaccount
type SubAccount struct {
	ID              string `json:"id"`
	SplitPercentage int    `json:"split_percentage"`
}

// Customization represents Flutterwave payment customization
type Customization struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Logo        string `json:"logo,omitempty"`
}

// FlutterwaveMobileMoneyRequest represents Mobile Money request for Ghana
type FlutterwaveMobileMoneyRequest struct {
	AccountNumber string `json:"account_number"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	TxRef         string `json:"tx_ref"`
	Network       string `json:"network"` // "MTN", "VODAFONE", "TIGO", "AIRTEL"
}

// FlutterwavePayoutRequest represents a payout request to Flutterwave for Ghana
type FlutterwavePayoutRequest struct {
	AccountBank     string `json:"account_bank"`
	AccountNumber   string `json:"account_number"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Narration       string `json:"narration"`
	Reference       string `json:"reference"`
	BeneficiaryName string `json:"beneficiary_name"`
}

// TransactionVerificationRequest represents a verification request
type TransactionVerificationRequest struct {
	ID string `json:"id"`
}
