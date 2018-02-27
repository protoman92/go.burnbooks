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
	Available() <-chan interface{}
	TakeResult() <-chan *BookTakeResult
}

type bookPile struct {
	available   chan interface{}
	books       chan Burnable
	id          string
	takeResult  chan *BookTakeResult
	takeTimeout time.Duration
}

func (bp *bookPile) Available() <-chan interface{} {
	return bp.available
}

func (bp *bookPile) Supply(taker BookTaker) {
	go func() {
		bookCh := bp.books
		capacity := taker.Capacity()
		loaded := make([]Burnable, 0)
		var availableCh, startLoadCh chan<- interface{}
		var delayTake chan interface{}
		var loadBookCh chan<- []Burnable
		var loadResult *BookTakeResult
		var takeResultCh chan<- *BookTakeResult

		for {
			var bookCh2 chan Burnable

			if len(loaded) < capacity {
				// Only when loaded has not filled capacity do we initialize the book
				// channel. Otherwise, this is a nil channel that will not be selected.
				bookCh2 = bookCh
			}

			// The sequence of operation here is:
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
			// - Once the availability channel has received a signal, reset the loaded
			// slice to start another loading process.
			//
			// A possible optimization is to send the take result in another goroutine
			// so as not to block the rest of the sequence.
			select {
			case book := <-bookCh2:
				loaded = append(loaded, book)

				if len(loaded) == capacity {
					// This must be initialized with 1 buffer slot so that it does not
					// block when we try to insert value below.
					startLoadCh = make(chan interface{}, 1)
				}

			case <-time.After(bp.takeTimeout):
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
				loadBookCh = taker.LoadBooks()

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

				takeResultCh = bp.takeResult

			case takeResultCh <- loadResult:
				takeResultCh = nil
				loadResult = nil

				// Buffer 1 to avoid blocking on delayTake.
				delayTake = make(chan interface{}, 1)

			case delayTake <- true:
				delayTake = nil

				if len(loaded) == capacity {
					// If the number of loaded books is equal to the taker's capacity,
					// there is a good chance that there are still books in this pile
					// (unless the initial number of books is exactly divisible by said
					// capacity, but this would only add one more loop). Therefore, we
					// send an availability signal.
					availableCh = bp.available
				}

				time.Sleep(taker.TakeDelay())

			case availableCh <- true:
				availableCh = nil

				// Reset the loaded slice here to enable next round of loading. This
				// is done at the last step of the process, before we force a delay on
				// the next take operation.
				loaded = make([]Burnable, 0)
			}
		}
	}()
}

func (bp *bookPile) TakeResult() <-chan *BookTakeResult {
	return bp.takeResult
}

// NewBookPile creates a new BookPile.
func NewBookPile(params *BookPileParams) FBookPile {
	books := params.books
	bookCh := make(chan Burnable, len(books))

	for _, book := range books {
		bookCh <- book
	}

	pile := &bookPile{
		available:   make(chan interface{}),
		books:       bookCh,
		id:          params.id,
		takeResult:  make(chan *BookTakeResult),
		takeTimeout: 1e9,
	}

	go func() {
		pile.available <- true
	}()

	return pile
}
