package goburnbooks

import (
	"sort"
	"strconv"
	"testing"
	"time"
)

const (
	defaultTimeout = 1e8
)

func newRandomBookPile(count int, offset int) FBookPile {
	books := make([]Burnable, count)

	for i := 0; i < count; i++ {
		id := offset + i
		params := &BookParams{BurnDuration: 0, ID: strconv.Itoa(id)}
		books[i] = NewBook(params)
	}

	bookParams := &BookPileParams{books: books}
	return NewBookPile(bookParams)
}

func Test_BookTakersHavingOddCapacity_ShouldStillLoadAllBooks(t *testing.T) {
	// Setup
	t.Parallel()
	pileCount, bookCount := 10, 10
	totalBookCount := pileCount * bookCount
	bookPiles := make([]FBookPile, pileCount)

	for ix := range bookPiles {
		pile := newRandomBookPile(bookCount, ix*bookCount)
		bookPiles[ix] = pile
	}

	pileGroup := NewBookPileGroup(bookPiles...)

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

	// This ensures that loaded books are unique.
	loadedBooks := make(map[Burnable]int, 0)
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
						loadedBooks[item] = loadedBooks[item] + 1
					}
				}
			}
		}
	}()

	// When
	for _, taker := range bookTakers {
		go func(taker BookTaker) {
			for {
				pileGroup.Supply(taker)
				time.Sleep(1e5)
			}
		}(taker)
	}

	time.Sleep(2e9)

	// Then
	loadedLen := len(loadedBooks)

	if loadedLen != totalBookCount {
		t.Errorf("Should have taken %d, instead got %d", totalBookCount, loadedLen)
	}

	keys := make([]int, 0)

	for key, value := range loadedBooks {
		numericKey, _ := strconv.Atoi(key.UID())
		keys = append(keys, numericKey)

		if value != 1 {
			t.Errorf("%v should have been taken once, but instead was %d", key, value)
		}
	}

	sort.Ints(keys)
	keyLen := len(keys)

	if keyLen != totalBookCount {
		t.Errorf("Should have %d keys, instead got %d", totalBookCount, keyLen)
	}
}
