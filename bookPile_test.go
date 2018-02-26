package goburnbooks

import (
	"strconv"
	"testing"
	"time"
)

const (
	defaultTimeout = 1e8
)

func newRandomBookPile(count int) BookPile {
	books := make([]Burnable, count)

	for i := 0; i < count; i++ {
		params := &BookParams{BurnDuration: 0, UID: strconv.Itoa(i)}
		books[i] = NewBook(params)
	}

	bookParams := &BookPileParams{books: books}
	return NewBookPile(bookParams)
}

func Test_BookTakersHavingOddCapacity_ShouldStillLoadAllBooks(t *testing.T) {
	// Setup
	t.Parallel()
	pileCount := 100000
	pile := newRandomBookPile(pileCount)
	takerCount := 100
	bookTakers := make([]BookTaker, takerCount)
	bookChs := make([]chan []Burnable, takerCount)

	for ix := range bookTakers {
		loadBookCh := make(chan []Burnable)

		btParams := &BookTakerParams{
			capacity:  17,
			loadBooks: loadBookCh,
			timeout:   1e9,
		}

		bookTaker := NewBookTaker(btParams)
		bookTakers[ix] = bookTaker
		bookChs[ix] = loadBookCh
	}

	loadedBooks := make(map[Burnable]bool, 0)
	updateLoaded := make(chan []Burnable)

	for _, ch := range bookChs {
		go func(ch chan []Burnable) {
			for {
				select {
				case burnables := <-ch:
					go func() {
						updateLoaded <- burnables
					}()
				}
			}
		}(ch)
	}

	go func() {
		for {
			select {
			case loaded := <-updateLoaded:
				if len(loaded) > 0 {
					for _, item := range loaded {
						loadedBooks[item] = true
					}
				}
			}
		}
	}()

	// When
	for _, taker := range bookTakers {
		go func(taker BookTaker) {
			for {
				pile.Supply(taker)
				time.Sleep(1e5)
			}
		}(taker)
	}

	time.Sleep(2e9)

	// Then
	loadedLen := len(loadedBooks)

	if loadedLen != pileCount {
		t.Errorf("Should have taken %d, instead got %d", pileCount, loadedLen)
	}
}
