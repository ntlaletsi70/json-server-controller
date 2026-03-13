package service

import (
	"encoding/json"
	"fmt"
)

func validateJSONPayload(raw string) error {

	if len(raw) == 0 {
		return fmt.Errorf("json payload cannot be empty")
	}

	if len(raw) > 1_000_000 {
		return fmt.Errorf("json payload too large (max 1MB)")
	}

	var parsed interface{}

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if _, ok := parsed.(map[string]interface{}); !ok {
		return fmt.Errorf("top level JSON must be an object")
	}

	return nil
}
