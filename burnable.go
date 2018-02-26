package goburnbooks

// Burnable represents something that can burn, e.g. books. In the Burn() method,
// we may implement variable sleep durations to simulate different burning
// processes (bigger books burn more slowly).
//
// For the sake of simplicity, we assume that everything can be burnt eventually,
// only that some do so longer than others. Therefore, the Burn() method does
// not error out.
type Burnable interface {
	UID() string
	Burn()
}

// BurnResultCollector collects burn results.
type BurnResultCollector interface {

	// Every time a Burnable is burned, put it here.
	BurnResult() <-chan *BurnResult
}

// BurnResult represents the result of a burning.
type BurnResult struct {
	incineratorID string
	burned        Burnable
}
