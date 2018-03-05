package goburnbooks

import (
	"fmt"
	"sync"
	"time"
)

// SupplyPile represents a pile of Suppliables.
type SupplyPile interface {
	Supply(taker SupplyTaker)
}

// FSupplyPile represents a SupplyPile that has all functionalities.
type FSupplyPile interface {
	SupplyPile
	TakeResultChannel() <-chan SupplyTakeResult
}

// SupplyPileParams represents the required parameters to build a SupplyPile.
// The take timeout should be a minimal value to prevent it from hogging the
// loading process when a pile has no more supply to offer (which happens quite
// frequently if there are many piles with small supply count). It should not
// be 0, however, because that will randomize the select sequence so much so
// that loading becomes suboptimal.
type SupplyPileParams struct {
	Logger             Logger
	Supply             []Suppliable
	ID                 string
	TakeResultCapacity uint
	TakeTimeout        time.Duration
}

type supplyPile struct {
	mutex sync.Mutex
	SupplyPileParams
	supplyCh     chan Suppliable
	takeResultCh chan SupplyTakeResult
}

func (sp *supplyPile) String() string {
	return fmt.Sprintf("Supply pile %s", sp.ID)
}

func (sp *supplyPile) Supply(taker SupplyTaker) {
	go func() {
		capacity := taker.Capacity()
		loaded := make([]Suppliable, 0)
		logger := sp.Logger
		readyCh := taker.SendTakeReadyChannel()
		takerID := taker.SupplyTakerID()
		var loadResult SupplyTakeResult
		var loadSupplyCh chan<- []Suppliable
		var resetSequenceCh chan interface{}
		var startLoadCh chan<- interface{}
		var supplyCh chan Suppliable
		var supplyTimeoutCh <-chan time.Time
		var takeResultCh chan SupplyTakeResult

		for {
			// The sequence of operation here is:
			// - The pile waits for the taker to be ready first, then initialize the
			// supplyCh and timeout channels. The ready channel is then nullified to
			// ignore subsequent requests.
			// - The supply channel and the timeout channel compete to emit. If the
			// supply channel is empty, the timeout channel will win eventually.
			// - Once the taker capacity has been reached, or the timeout happens,
			// signal that take can happen.
			// - Once start load happens, signal that loading can happen, but only
			// if there is a positive number of loaded items. Otherwise, signal ready
			// and wait for the next request.
			// - Once all the supplies have been loaded, initialize the result channel.
			// - After the result has been deposited, initialize the reset channel.
			// - Finally, reset the ready channel and the loaded slice to prepare for
			// another loading process
			select {
			case <-readyCh:
				// Nullify the ready channel here to let the sequence run in peace.
				logger.Printf("%v: received ready from %v", sp, taker)
				readyCh = nil
				supplyCh = sp.supplyCh
				supplyTimeoutCh = time.After(sp.TakeTimeout)

			// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			case supply, ok := <-supplyCh:
				if !ok {
					return
				}

				loaded = append(loaded, supply)

				if uint(len(loaded)) == capacity {
					logger.Printf("%v: supplied to full (%d) for %v", sp, capacity, taker)
					supplyCh = nil
					supplyTimeoutCh = nil

					// This must be initialized with 1 buffer slot so that it does not
					// block when we try to insert value below.
					startLoadCh = make(chan interface{}, 1)
				}

			case <-supplyTimeoutCh:
				logger.Printf("%v: timed out for %v!", sp, taker)
				supplyTimeoutCh = nil
				supplyCh = nil
				startLoadCh = make(chan interface{}, 1)
			// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

			case startLoadCh <- true:
				// With this setup, we can be sure that once this stage this reached,
				// the above two select cases will not be selected again:
				// - The supply channel will always be nil because we have loaded to the
				// taker's full capacity.
				// - The timeout channel will not emit an element fast enough to be
				// selected.
				startLoadCh = nil

				if len(loaded) > 0 {
					// Only initialize the load supply channel when there are loaded items.
					// Beware that if the taker relies on this channel to orchestrate
					// its work, said taker should have some mechanism to detect lack of
					// signal in order to send its requests elsewhere, such as timeout.
					loadSupplyCh = taker.ReceiveLoadChannel()
				} else {
					logger.Printf("%v: did not supply anything for %v", sp, taker)
					resetSequenceCh = make(chan interface{}, 1)
				}

			case loadSupplyCh <- loaded:
				logger.Printf("%v: supplied %d to %v", sp, len(loaded), taker)
				loadSupplyCh = nil
				takeResultCh = sp.takeResultCh

				supplyIds := make([]string, len(loaded))

				for ix, supply := range loaded {
					supplyIds[ix] = supply.SuppliableID()
				}

				loadResult = NewTakeResult(sp.ID, takerID, supplyIds)

			case takeResultCh <- loadResult:
				takeResultCh = nil
				loadResult = nil
				resetSequenceCh = make(chan interface{}, 1)

			case resetSequenceCh <- true:
				resetSequenceCh = nil
				loaded = make([]Suppliable, 0)
				readyCh = taker.SendTakeReadyChannel()
			}
		}
	}()
}

func (sp *supplyPile) TakeResultChannel() <-chan SupplyTakeResult {
	return sp.takeResultCh
}

// NewSupplyPile creates a new SupplyPile.
func NewSupplyPile(params *SupplyPileParams) FSupplyPile {
	supplies := params.Supply
	supplyCh := make(chan Suppliable, len(supplies))

	for _, supply := range supplies {
		supplyCh <- supply
	}

	pile := &supplyPile{
		SupplyPileParams: *params,
		supplyCh:         supplyCh,
		takeResultCh:     make(chan SupplyTakeResult, params.TakeResultCapacity),
	}

	return pile
}
