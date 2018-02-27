package goburnbooks

import (
	"time"
)

// BookPile represents a pile of Books.
type BookPile interface {
	Supply(taker BookTaker)
}

// BookPileParams represents the required parameters to build a BookPile.
type BookPileParams struct {
	books []Burnable
	id    string
}

// FBookPile represents a pile of Books with all functionalities.
type FBookPile interface {
	BookPile
	TakeResultChannel() <-chan *BookTakeResult
}

type bookPile struct {
	bookCh       chan Burnable
	id           string
	takeResultCh chan *BookTakeResult
	takeTimeout  time.Duration
}

func (bp *bookPile) Supply(taker BookTaker) {
	go func() {
		capacity := taker.Capacity()
		loaded := make([]Burnable, 0)
		readyCh := taker.ReadyChannel()
		var bookCh chan Burnable
		var loadBookCh chan<- []Burnable
		var loadResult *BookTakeResult
		var startLoadCh chan<- interface{}
		var takeResultCh chan<- *BookTakeResult
		var timeoutCh <-chan time.Time

		for {
			// The sequence of operation here is:
			// - The pile waits for the taker to be ready first, then initialize the
			// book and timeout channels. The ready channel is then nullified to
			// ignore subsequent requests.
			// - The book channel and the timeout channel compete to emit. If the
			// book channel is empty, the timeout channel will win eventually.
			// - Once the taker capacity has been reached, or the timeout happens,
			// signal that take can happen.
			// - Once start load happens, signal that loading can happen.
			// - Once all the books have been loaded, initialize the load result and
			// result channel, and signal that the result can be deposited.
			// - Once the result has been consumed, signal the delay channel to sleep
			// for a specified period of time.
			// - Once the delay channel has completed work, signal the availability
			// channel.
			// - Once the availability channel has received a signal, reset the ready
			// channel and the loaded slice to start another loading process.
			//
			// A possible optimization is to send the take result in another goroutine
			// so as not to block the rest of the sequence.
			select {
			case <-readyCh:
				// Nullify the ready channel here to let the sequence run in peace.
				readyCh = nil
				bookCh = bp.bookCh
				timeoutCh = time.After(bp.takeTimeout)

			case book := <-bookCh:
				loaded = append(loaded, book)

				if len(loaded) == capacity {
					bookCh = nil

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
				// - The book channel will always be nil because we have loaded to the
				// taker's full capacity.
				// - The timeout channel will not emit an element fast enough to be
				// selected.
				startLoadCh = nil

				// If the available channel was initialized, these two channels will
				// essentially be chosen at random. It does not matter which one goes
				// first, however.
				loadBookCh = taker.LoadChannel()

			case loadBookCh <- loaded:
				loadBookCh = nil
				bookIds := make([]string, len(loaded))

				for ix, book := range loaded {
					bookIds[ix] = book.UID()
				}

				// Only at this step do both of the variables below get set. When the
				// result has been successfully deposited, deinitialize them immediately.
				loadResult = &BookTakeResult{
					bookIds: bookIds,
					pileID:  bp.id,
					takerID: taker.UID(),
				}

				takeResultCh = bp.takeResultCh

			case takeResultCh <- loadResult:
				if len(loaded) != capacity {
					// If the number of loaded books is not equal to the taker's capacity,
					// the pile does not have enough books left for another take operation.
					return
				}

				takeResultCh = nil
				loadResult = nil

				// Reset the loaded slice here to enable next round of loading. This
				// is done at the last step of the process, before we force a delay on
				// the next take operation.
				loaded = make([]Burnable, 0)

				// Reinstate the ready channel to start taking requests again.
				readyCh = taker.ReadyChannel()
			}
		}
	}()
}

func (bp *bookPile) TakeResultChannel() <-chan *BookTakeResult {
	return bp.takeResultCh
}

// NewBookPile creates a new BookPile.
func NewBookPile(params *BookPileParams) FBookPile {
	books := params.books
	bookCh := make(chan Burnable, len(books))

	for _, book := range books {
		bookCh <- book
	}

	pile := &bookPile{
		bookCh:       bookCh,
		id:           params.id,
		takeResultCh: make(chan *BookTakeResult),
		takeTimeout:  1e9,
	}

	return pile
}
