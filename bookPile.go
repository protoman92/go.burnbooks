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
}

// FBookPile represents a pile of Books with all functionalities.
type FBookPile interface {
	BookPile
	Available() <-chan interface{}
}

type bookPile struct {
	available chan interface{}
	books     chan Burnable
}

func (bp *bookPile) Available() <-chan interface{} {
	return bp.available
}

func (bp *bookPile) Supply(taker BookTaker) {
	go func() {
		bookCh := bp.books
		capacity := taker.Capacity()
		loaded := make([]Burnable, 0)
		startLoad := make(chan interface{})

		for {
			var bookCh2 chan Burnable

			if len(loaded) < capacity {
				// Only when loaded has not filled capacity do we initialize the book
				// channel. Otherwise, this is a nil channel that will not be selected.
				bookCh2 = bookCh
			}

			select {
			case book := <-bookCh2:
				loaded = append(loaded, book)

				if len(loaded) == capacity {
					go func() {
						startLoad <- true
					}()
				}

			case <-time.After(taker.TakeTimeout()):
				go func() {
					startLoad <- true
				}()

			case <-startLoad:
				if len(loaded) == capacity {
					// If the number of loaded books is equal to the taker's capacity,
					// there is a good chance that there are still books in this pile
					// (unless the initial number of books is exactly divisible by said
					// capacity, but this would only add one more loop). Therefore, we
					// send an availability signal.
					bp.available <- true
				}

				go func() {
					taker.LoadBooks() <- loaded
				}()

				return
			}
		}
	}()
}

// NewBookPile creates a new BookPile.
func NewBookPile(params *BookPileParams) FBookPile {
	books := params.books
	bookCh := make(chan Burnable, len(books))

	for _, book := range books {
		bookCh <- book
	}

	pile := &bookPile{available: make(chan interface{}), books: bookCh}

	go func() {
		pile.available <- true
	}()

	return pile
}
