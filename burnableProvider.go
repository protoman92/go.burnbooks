package goburnbooks

// BurnableProvider represents a Burnable provider.
type BurnableProvider interface {
	ProvideChannel() <-chan []Burnable

	// Beware that an emission does not mean the previous work load has been
	// finished, just that the incinerator has burned enough to take in more, and
	// it does not expect the remaining load to take long.
	ReadyChannel() chan<- interface{}
	UID() string
}

// BurnableProviderParams represents all the required parameters to build a
// provider.
type BurnableProviderParams struct {
	ID        string
	ProvideCh <-chan []Burnable
	ReadyCh   chan<- interface{}
}

type burnableProvider struct {
	*BurnableProviderParams
}

func (bp *burnableProvider) ProvideChannel() <-chan []Burnable {
	return bp.ProvideCh
}

func (bp *burnableProvider) ReadyChannel() chan<- interface{} {
	return bp.ReadyCh
}

func (bp *burnableProvider) UID() string {
	return bp.ID
}

// NewBurnableProvider returns a new BurnableProvider.
func NewBurnableProvider(params *BurnableProviderParams) BurnableProvider {
	return &burnableProvider{BurnableProviderParams: params}
}
