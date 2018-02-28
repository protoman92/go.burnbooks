package goburnbooks

import (
	"fmt"
)

// Burnable represents something that can burn, e.g. books. In the Burn() method,
// we may implement variable sleep durations to simulate different burning
// processes (bigger books burn more slowly).
//
// For the sake of simplicity, we assume that everything can be burnt eventually,
// only that some do so longer than others. Therefore, the Burn() method does
// not error out.
type Burnable interface {
	BurnableID() string
	Burn()
}

// BurnResult represents the result of a burning.
type BurnResult struct {
	Burned        Burnable
	IncineratorID string
	ProviderID    string
}

func (br *BurnResult) String() string {
	return fmt.Sprintf(
		"Burned %v with incinerator %s, provided by %s",
		br.Burned,
		br.IncineratorID,
		br.ProviderID,
	)
}
