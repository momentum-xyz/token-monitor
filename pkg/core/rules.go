package core

import (
	"encoding/json"
	"math/big"

	"github.com/ory/x/errorsx"
)

type RuleDefinition struct {
	ID           int                       `json:"id"`
	Active       bool                      `json:"active"`
	Token        TokenDefinition           `json:"token"`
	Network      string                    `json:"network"`
	Requirements RuleRequirementDefinition `json:"requirements"`
}

type TokenDefinition struct {
	Type    string   `json:"type"`
	Address string   `json:"address"`
	TokenID *big.Int `json:"token_id"`
}

type RuleRequirementDefinition struct {
	MinBalance *big.Int `json:"minimum_balance"`
}

func UnmarshalRuleDefinition(ruleJSON []byte) (*RuleDefinition, error) {
	var rule *RuleDefinition
	err := json.Unmarshal(ruleJSON, rule)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}
	return rule, nil
}

func UnmarshalRuleDefinitions(rulesJsonString []byte) ([]*RuleDefinition, error) {
	var rules []*RuleDefinition
	err := json.Unmarshal(rulesJsonString, &rules)
	if err != nil {
		return nil, errorsx.WithStack(err)
	}
	return rules, nil
}
