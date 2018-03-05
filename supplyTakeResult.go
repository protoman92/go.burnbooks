package goburnbooks

import (
	"fmt"
)

// SupplyTakeResult represents the result of a take operation.
type SupplyTakeResult interface {
	PileID() string
	TakerID() string
	SupplyIDs() []string
}

type supplyTakeResult struct {
	pileID    string
	takerID   string
	supplyIDs []string
}

func (str *supplyTakeResult) PileID() string {
	return str.pileID
}

func (str *supplyTakeResult) TakerID() string {
	return str.takerID
}

func (str *supplyTakeResult) SupplyIDs() []string {
	return str.supplyIDs
}

func (str *supplyTakeResult) String() string {
	return fmt.Sprintf(
		"Supply taker %s took %d supplies from pile %s",
		str.takerID,
		len(str.supplyIDs),
		str.pileID,
	)
}

// NewTakeResult returns a new SupplyTakeResult.
func NewTakeResult(pileID string, takerID string, supplyIDs []string) SupplyTakeResult {
	return &supplyTakeResult{
		pileID:    pileID,
		takerID:   takerID,
		supplyIDs: supplyIDs,
	}
}
