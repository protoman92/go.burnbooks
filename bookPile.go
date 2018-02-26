package goburnbooks

import (
	"time"
)

// BookPile represents a pile of Books.
type BookPile interface {
	Books() <-chan Burnable
	Supply(taker BookTaker)
}

// BookPileParams represents the required parameters to build a BookPile.
type BookPileParams struct {
	books []Burnable
}

type bookPile struct {
	books chan Burnable
}

func (bp *bookPile) Books() <-chan Burnable {
	return bp.books
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
				go func() {
					taker.LoadBooks() <- loaded
				}()

				return
			}
		}
	}()
}

// NewBookPile creates a new BookPile.
func NewBookPile(params *BookPileParams) BookPile {
	books := params.books
	bookCh := make(chan Burnable, len(books))

	for _, book := range books {
		bookCh <- book
	}

	return &bookPile{books: bookCh}
}
