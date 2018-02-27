package goburnbooks

// BurnableProvider represents a Burnable provider.
type BurnableProvider interface {
	BurnableProviderID() string

	// Beware that an emission does not mean the previous work load has been
	// finished, just that the incinerator has burned enough to take in more, and
	// it does not expect the remaining load to take long.
	ConsumeReadyChannel() chan<- interface{}

	ProvideChannel() <-chan []Burnable
}

// BurnableProviderRawParams represents only the immutable parameters used to
// build a provider.
type BurnableProviderRawParams struct {
	BPID string
}

// BurnableProviderParams represents all the required parameters to build a
// provider.
type BurnableProviderParams struct {
	*BurnableProviderRawParams
	ConsumeReadyCh chan<- interface{}
	ProvideCh      <-chan []Burnable
}

type burnableProvider struct {
	*BurnableProviderParams
}

func (bp *burnableProvider) ProvideChannel() <-chan []Burnable {
	return bp.ProvideCh
}

func (bp *burnableProvider) ConsumeReadyChannel() chan<- interface{} {
	return bp.ConsumeReadyCh
}

func (bp *burnableProvider) BurnableProviderID() string {
	return bp.BPID
}

// NewBurnableProvider returns a new BurnableProvider.
func NewBurnableProvider(params *BurnableProviderParams) BurnableProvider {
	return &burnableProvider{BurnableProviderParams: params}
}
