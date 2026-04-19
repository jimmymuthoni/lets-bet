package kyc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/config"
)

// ErrKYCNotConfigured is returned when Smile ID credentials are missing.
var ErrKYCNotConfigured = errors.New("smile id is not configured")

// Provider is the KYC provider interface used by the wallet/user services.
type Provider interface {
	VerifyUser(ctx context.Context, user *domain.User, idType, idNumber string) (*VerifyResult, error)
	VerifyID(ctx context.Context, countryCode, idType, idNumber string) (*VerifyResult, error)
}

// VerifyResult is the normalized KYC verification result.
type VerifyResult struct {
	Verified     bool   `json:"verified"`
	Status       string `json:"status"`
	FullName     string `json:"full_name,omitempty"`
	IDNumber     string `json:"id_number,omitempty"`
	IDType       string `json:"id_type,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SmileIDProvider calls the Smile ID REST API directly. We do not depend on
// the SDK types because they live under internal/ packages.
type SmileIDProvider struct {
	apiKey    string
	partnerID string
	baseURL   string
	http      *http.Client
}

// NewSmileIDProvider builds a Smile ID provider from config.
func NewSmileIDProvider(cfg config.SmileIDConfig) (*SmileIDProvider, error) {
	if cfg.APIKey == "" || cfg.PartnerID == "" {
		return nil, ErrKYCNotConfigured
	}
	baseURL := "https://api.smileidentity.com/v1"
	if strings.EqualFold(cfg.Env, "sandbox") {
		baseURL = "https://testapi.smileidentity.com/v1"
	}
	return &SmileIDProvider{
		apiKey:    cfg.APIKey,
		partnerID: cfg.PartnerID,
		baseURL:   baseURL,
		http:      &http.Client{Timeout: 15 * time.Second},
	}, nil
}

type kycRequest struct {
	PartnerID   string `json:"partner_id"`
	CountryCode string `json:"country_code"`
	IDType      string `json:"id_type"`
	IDNumber    string `json:"id_number"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
}

type kycResponse struct {
	Success      bool   `json:"success"`
	Verified     bool   `json:"verified"`
	FullName     string `json:"full_name"`
	IDNumber     string `json:"id_number"`
	IDType       string `json:"id_type"`
	CountryCode  string `json:"country_code"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// VerifyUser runs a digital KYC check for the given user using their ID document.
func (p *SmileIDProvider) VerifyUser(ctx context.Context, user *domain.User, idType, idNumber string) (*VerifyResult, error) {
	body := kycRequest{
		PartnerID:   p.partnerID,
		CountryCode: user.CountryCode,
		IDType:      idType,
		IDNumber:    idNumber,
		FirstName:   firstName(user.FullName),
		LastName:    lastName(user.FullName),
	}
	resp, err := p.post(ctx, "/kyc/verify", body)
	if err != nil {
		return nil, err
	}
	return toResult(resp), nil
}

// VerifyID runs an identity verification check for a national ID / passport.
func (p *SmileIDProvider) VerifyID(ctx context.Context, countryCode, idType, idNumber string) (*VerifyResult, error) {
	body := kycRequest{
		PartnerID:   p.partnerID,
		CountryCode: countryCode,
		IDType:      idType,
		IDNumber:    idNumber,
	}
	resp, err := p.post(ctx, "/identity/verify", body)
	if err != nil {
		return nil, err
	}
	return toResult(resp), nil
}

func (p *SmileIDProvider) post(ctx context.Context, path string, body any) (*kycResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", p.apiKey)
	req.Header.Set("X-Partner-Id", p.partnerID)

	resp, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("smile id request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("smile id api error: %s", resp.Status)
	}

	var decoded kycResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &decoded, nil
}

func toResult(r *kycResponse) *VerifyResult {
	status := "failed"
	if r.Verified {
		status = "verified"
	}
	return &VerifyResult{
		Verified:     r.Verified,
		Status:       status,
		FullName:     r.FullName,
		IDNumber:     r.IDNumber,
		IDType:       r.IDType,
		CountryCode:  r.CountryCode,
		ErrorMessage: r.ErrorMessage,
	}
}

func firstName(full string) string {
	if i := strings.IndexByte(full, ' '); i > 0 {
		return full[:i]
	}
	return full
}

func lastName(full string) string {
	if i := strings.LastIndexByte(full, ' '); i > 0 {
		return full[i+1:]
	}
	return ""
}
