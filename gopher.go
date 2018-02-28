package goburnbooks

import (
	"fmt"
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
	Logger       Logger
	TakeTimeout  time.Duration
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
	return fmt.Sprintf("Gopher %s", g.BPID)
}

func (g *gopher) loopWork() {
	logger := g.Logger
	resetSequenceCh := make(chan interface{}, 1)
	takeReadyCh := g.stTakeReadyCh
	var burnables []Burnable
	var burnableProvideCh chan []Burnable
	var consumeReadyCh chan interface{}
	var supplyLoadCh chan []Suppliable
	var takeTimeoutCh <-chan time.Time

	for {
		// The sequence of work is as follows:
		// - Firstly, signal that this gopher is ready to take some supplies, then
		// immediately set the take ready channel to nil to discard it in the next
		// loop. Initialize the supply load to start receiving supplies, and the
		// timeout channel to reset the process if no supplies arrive on time.
		// - When supplies arrive, nullify the supply load channel. Extract the
		// Burnables from said supplies, then sleep for a while to simulate trip
		// duration. Afterwards, initialize the provide channel to feed Burnables
		// downstream.
		// - Alternatively, if the supply did not arrive on time, hit the time out
		// and send a reset request.
		// - Once the Burnables have been fed downstream, nullify the Burnables
		// and provide channel and initialize the consume ready channel to wait
		// for availability signal.
		// - Once availability signal is retrieved, nullify the consume ready and
		// send a reset signal.
		// - Finally, reinstate the take ready channel to wait for the next batch.
		//
		select {
		case takeReadyCh <- true:
			takeReadyCh = nil
			supplyLoadCh = g.stLoadCh
			takeTimeoutCh = time.After(g.TakeTimeout)

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		case supplies := <-supplyLoadCh:
			logger.Printf("%v received %d supplies", g, len(supplies))
			supplyLoadCh = nil
			takeTimeoutCh = nil
			burnables = ExtractBurnablesFromSuppliables(supplies...)
			time.Sleep(g.TripDuration)
			burnableProvideCh = g.bpProvideCh

		case <-takeTimeoutCh:
			logger.Printf("%v timed out!", g)
			takeTimeoutCh = nil
			supplyLoadCh = nil
			resetSequenceCh <- true
		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

		case burnableProvideCh <- burnables:
			burnableProvideCh = nil
			burnables = nil
			consumeReadyCh = g.bpConsumerReadyCh

		case <-consumeReadyCh:
			consumeReadyCh = nil
			resetSequenceCh <- true

		case <-resetSequenceCh:
			takeReadyCh = g.stTakeReadyCh
		}
	}
}

// NewGopher returns a new Gopher.
func NewGopher(params *GopherParams) Gopher {
	bpRawParams := params.BurnableProviderRawParams
	stRawParams := params.SupplyTakerRawParams
	bpConsumeReadyCh := make(chan interface{})
	bpProvideCh := make(chan []Burnable)
	stLoadCh := make(chan []Suppliable)
	stTakeReadyCh := make(chan interface{})

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
