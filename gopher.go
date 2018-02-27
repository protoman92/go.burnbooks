package goburnbooks

import (
	"time"
)

// Gopher represents a worker in the system.
type Gopher interface {
	BurnableProvider
	SupplyTaker
}

// GopherParams represents all the required parameters to build a Gopher.
type GopherParams struct {
	*BurnableProviderRawParams
	*SupplyTakerRawParams
	TripDuration time.Duration
}

type gopher struct {
	BurnableProvider
	SupplyTaker
	*GopherParams
	bpConsumerReadyCh chan interface{}
	bpProvideCh       chan []Burnable
	stLoadCh          chan []Suppliable
	stTakeReadyCh     chan interface{}
}

func (g *gopher) String() string {
	return g.BPID
}

func (g *gopher) loopWork() {
	takeReadyCh := g.stTakeReadyCh
	var burnables []Burnable
	var burnableProvideCh chan []Burnable
	var consumeReadyCh chan interface{}
	var supplyLoadCh chan []Suppliable

	for {
		// The sequence of work is as follows:
		// - Firstly, signal that this gopher is ready to take some supplies, then
		// immediately set the take ready channel to nil to discard it in the next
		// loop. Initialize the supply load to start receiving supplies.
		// - When supplies arrive, nullify the supply load channel. Extract the
		// Burnables from said supplies, then sleep for a while to simulate trip
		// duration. Afterwards, initialize the provide channel to feed Burnables
		// downstream.
		// - Once the Burnables have been fed downstream, nullify the Burnables
		// and provide channel and initialize the consume ready channel to wait
		// for availability signal.
		// - Once availability signal is retrieved, nullify the consume ready
		// channel and reinstate the take ready channel to wait for the next batch.
		//
		select {
		case takeReadyCh <- true:
			takeReadyCh = nil
			supplyLoadCh = g.stLoadCh

		case supplies := <-supplyLoadCh:
			supplyLoadCh = nil
			burnables = ExtractBurnablesFromSuppliables(supplies...)
			time.Sleep(g.TripDuration)
			burnableProvideCh = g.bpProvideCh

		case burnableProvideCh <- burnables:
			burnableProvideCh = nil
			burnables = nil
			consumeReadyCh = g.bpConsumerReadyCh

		case <-consumeReadyCh:
			consumeReadyCh = nil
			takeReadyCh = g.stTakeReadyCh
		}
	}
}

// NewGopher returns a new Gopher.
func NewGopher(params *GopherParams) Gopher {
	bpRawParams := params.BurnableProviderRawParams
	stRawParams := params.SupplyTakerRawParams
	bpConsumeReadyCh := make(chan interface{})
	bpProvideCh := make(chan []Burnable, 1)
	stLoadCh := make(chan []Suppliable)
	stTakeReadyCh := make(chan interface{}, 1)

	gp := &gopher{
		BurnableProvider: NewBurnableProvider(&BurnableProviderParams{
			BurnableProviderRawParams: bpRawParams,
			ConsumeReadyCh:            bpConsumeReadyCh,
			ProvideCh:                 bpProvideCh,
		}),
		SupplyTaker: NewSupplyTaker(&SupplyTakerParams{
			SupplyTakerRawParams: stRawParams,
			LoadCh:               stLoadCh,
			TakeReadyCh:          stTakeReadyCh,
		}),
		GopherParams:      params,
		bpConsumerReadyCh: bpConsumeReadyCh,
		bpProvideCh:       bpProvideCh,
		stLoadCh:          stLoadCh,
		stTakeReadyCh:     stTakeReadyCh,
	}

	go gp.loopWork()
	return gp
}
