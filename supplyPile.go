package goburnbooks

import (
	"time"
)

// SupplyPile represents a pile of Suppliables.
type SupplyPile interface {
	Supply(taker SupplyTaker)
}

// SupplyPileParams represents the required parameters to build a SupplyPile.
type SupplyPileParams struct {
	supply []Suppliable
	id     string
}

// FSupplyPile represents a pile of SupplyPiles with all functionalities.
type FSupplyPile interface {
	SupplyPile
	TakeResultChannel() <-chan *SupplyTakeResult
}

type supplyPile struct {
	supplyCh     chan Suppliable
	id           string
	takeResultCh chan *SupplyTakeResult
	takeTimeout  time.Duration
}

func (bp *supplyPile) Supply(taker SupplyTaker) {
	go func() {
		capacity := taker.Capacity()
		loaded := make([]Suppliable, 0)
		readyCh := taker.ReadyChannel()
		var loadResult *SupplyTakeResult
		var loadSupplyCh chan<- []Suppliable
		var startLoadCh chan<- interface{}
		var supplyCh chan Suppliable
		var takeResultCh chan<- *SupplyTakeResult
		var timeoutCh <-chan time.Time

		for {
			// The sequence of operation here is:
			// - The pile waits for the taker to be ready first, then initialize the
			// supplyCh and timeout channels. The ready channel is then nullified to
			// ignore subsequent requests.
			// - The supply channel and the timeout channel compete to emit. If the
			// supply channel is empty, the timeout channel will win eventually.
			// - Once the taker capacity has been reached, or the timeout happens,
			// signal that take can happen.
			// - Once start load happens, signal that loading can happen.
			// - Once all the supplies have been loaded, initialize the load result
			// and result channel, and signal that the result can be deposited.
			// - Once the result has been consumed, reset the ready channel and the
			// loaded slice to start another loading process.
			//
			// A possible optimization is to send the take result in another goroutine
			// so as not to block the rest of the sequence.
			select {
			case <-readyCh:
				// Nullify the ready channel here to let the sequence run in peace.
				readyCh = nil
				supplyCh = bp.supplyCh
				timeoutCh = time.After(bp.takeTimeout)

			case supply := <-supplyCh:
				loaded = append(loaded, supply)

				if len(loaded) == capacity {
					supplyCh = nil

					// This must be initialized with 1 buffer slot so that it does not
					// block when we try to insert value below.
					startLoadCh = make(chan interface{}, 1)
				}

			case <-timeoutCh:
				timeoutCh = nil
				startLoadCh = make(chan interface{}, 1)

			case startLoadCh <- true:
				// With this setup, we can be sure that once this stage this reached,
				// the above two select cases will not be selected again:
				// - The supply channel will always be nil because we have loaded to the
				// taker's full capacity.
				// - The timeout channel will not emit an element fast enough to be
				// selected.
				startLoadCh = nil

				// If the available channel was initialized, these two channels will
				// essentially be chosen at random. It does not matter which one goes
				// first, however.
				loadSupplyCh = taker.LoadChannel()

			case loadSupplyCh <- loaded:
				loadSupplyCh = nil
				supplyIds := make([]string, len(loaded))

				for ix, supply := range loaded {
					supplyIds[ix] = supply.UID()
				}

				// Only at this step do both of the variables below get set. When the
				// result has been successfully deposited, deinitialize them immediately.
				loadResult = &SupplyTakeResult{
					supplyIds: supplyIds,
					pileID:    bp.id,
					takerID:   taker.UID(),
				}

				takeResultCh = bp.takeResultCh

			case takeResultCh <- loadResult:
				if len(loaded) != capacity {
					// If the number of loaded supplies is not equal to the taker's
					// capacity, the pile does not have enough supplies left for another
					// take operation.
					return
				}

				takeResultCh = nil
				loadResult = nil

				// Reset the loaded slice here to enable next round of loading. This
				// is done at the last step of the process, before we force a delay on
				// the next take operation.
				loaded = make([]Suppliable, 0)

				// Reinstate the ready channel to start taking requests again.
				readyCh = taker.ReadyChannel()
			}
		}
	}()
}

func (bp *supplyPile) TakeResultChannel() <-chan *SupplyTakeResult {
	return bp.takeResultCh
}

// NewSupplyPile creates a new SupplyPile.
func NewSupplyPile(params *SupplyPileParams) FSupplyPile {
	supplies := params.supply
	supplyCh := make(chan Suppliable, len(supplies))

	for _, supply := range supplies {
		supplyCh <- supply
	}

	pile := &supplyPile{
		supplyCh:     supplyCh,
		id:           params.id,
		takeResultCh: make(chan *SupplyTakeResult),
		takeTimeout:  1e9,
	}

	return pile
}
