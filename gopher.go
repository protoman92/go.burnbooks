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
	BurnableProviderRawParams
	SupplyTakerRawParams
	Logger       Logger
	TripDuration time.Duration
}

type gopher struct {
	BurnableProvider
	SupplyTaker
	GopherParams
	receiveSupplyCh chan []Suppliable
	sendBurnableCh  chan []Burnable
}

func (g *gopher) String() string {
	return fmt.Sprintf("Gopher %s", g.BPID)
}

func (g *gopher) loopWork() {
	logger := g.Logger
	receiveSupplyCh := g.receiveSupplyCh
	var burnables []Burnable
	var sendBurnableCh chan []Burnable

	for {
		// Note that the logic in the gopher is quite simple. This is because the
		// heavy lifting has been delegated to the taker and provider. As a result
		// the gopher is only responsible for transfering resources from the receive
		// channel to the send channel and simulating travel time.
		select {
		case supplies := <-receiveSupplyCh:
			logger.Printf("%v received %d supplies", g, len(supplies))
			receiveSupplyCh = nil
			burnables = ExtractBurnablesFromSuppliables(supplies...)
			sendBurnableCh = g.sendBurnableCh
			time.Sleep(g.TripDuration)

		case sendBurnableCh <- burnables:
			sendBurnableCh = nil
			burnables = nil
			receiveSupplyCh = g.receiveSupplyCh
		}
	}
}

// NewGopher returns a new Gopher.
func NewGopher(params *GopherParams) Gopher {
	bpRawParams := params.BurnableProviderRawParams
	stRawParams := params.SupplyTakerRawParams
	receiveSupplyCh := make(chan []Suppliable)
	sendBurnablesCh := make(chan []Burnable)

	gp := &gopher{
		BurnableProvider: NewBurnableProvider(&BurnableProviderParams{
			BurnableProviderRawParams: bpRawParams,
			BPLogger:                  params.Logger,
			ReceiveBurnableSourceCh:   sendBurnablesCh,
		}),
		SupplyTaker: NewSupplyTaker(&SupplyTakerParams{
			SendSupplyDestCh:     receiveSupplyCh,
			SupplyTakerRawParams: stRawParams,
			STLogger:             params.Logger,
		}),
		GopherParams:    *params,
		receiveSupplyCh: receiveSupplyCh,
		sendBurnableCh:  sendBurnablesCh,
	}

	go gp.loopWork()
	return gp
}
