package goburnbooks

import "fmt"

// BurnResult represents the result of a burning.
type BurnResult interface {
	Burned() Burnable
	IncineratorID() string
	ProviderID() string
}

type burnResult struct {
	burned        Burnable
	incineratorID string
	providerID    string
}

func (br *burnResult) String() string {
	return fmt.Sprintf(
		"Burned %v with incinerator %s, provided by %s",
		br.burned,
		br.incineratorID,
		br.providerID,
	)
}

func (br *burnResult) Burned() Burnable {
	return br.burned
}

func (br *burnResult) IncineratorID() string {
	return br.incineratorID
}

func (br *burnResult) ProviderID() string {
	return br.providerID
}

// NewBurnResult returns a new BurnResult.
func NewBurnResult(burned Burnable, incID string, provID string) BurnResult {
	return &burnResult{burned: burned, incineratorID: incID, providerID: provID}
}
