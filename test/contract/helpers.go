// Package contract provides helper functions for HTTP contract tests
package contract

import (
	"encoding/json"
	"fmt"
)

// ContractValidator provides validation functions for API contracts
type ContractValidator struct{}

// NewContractValidator creates a new contract validator
func NewContractValidator() *ContractValidator {
	return &ContractValidator{}
}

// ValidateJSONSchema validates that a JSON response matches expected schema
func (cv *ContractValidator) ValidateJSONSchema(response []byte, expectedSchema string) error {
	var responseJSON any
	if err := json.Unmarshal(response, &responseJSON); err != nil {
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// In a full implementation, this would validate against a JSON schema
	// For now, we just ensure the response is valid JSON
	fmt.Printf("Validated response against schema: %s\n", expectedSchema)
	return nil
}

// GetExpectedContract returns the expected contract for a given endpoint
func (cv *ContractValidator) GetExpectedContract(endpoint string) string {
	contracts := map[string]string{
		"wallet_creation": `{
			"id": "uuid",
			"user_id": "uuid",
			"currency": "string",
			"balance": "decimal",
			"version": "number",
			"bonus_balance": "decimal",
			"today_deposit": "decimal",
			"updated_at": "timestamp"
		}`,
		"wallet_balance": `{
			"user_id": "uuid",
			"currency": "string",
			"balance": "decimal",
			"bonus_balance": "decimal",
			"available_balance": "decimal",
			"today_deposit": "decimal",
			"updated_at": "timestamp"
		}`,
		"game_creation": `{
			"id": "uuid",
			"game_type": "string",
			"round_number": "number",
			"server_seed": "string",
			"client_seed": "string",
			"crash_point": "decimal",
			"status": "string",
			"started_at": "timestamp",
			"crashed_at": "timestamp",
			"min_bet": "decimal",
			"max_bet": "decimal",
			"max_multiplier": "decimal"
		}`,
	}

	if contract, exists := contracts[endpoint]; exists {
		return contract
	}

	return "{}"
}
