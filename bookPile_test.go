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
	/// Setup
	t.Parallel()
	pileCount, bookCount := 20, 1000
	totalBCount := pileCount * bookCount
	bookPiles := make([]FBookPile, pileCount)
	t.Logf("Have %d books in total", totalBCount)

	for ix := range bookPiles {
		pile := newRandomBookPile(bookCount, ix*bookCount)
		bookPiles[ix] = pile
	}

	pileGroup := NewBookPileGroup(bookPiles...)
	takerCount := 50
	bookTakers := make([]BookTaker, takerCount)
	bookChs := make([]chan []Burnable, takerCount)

	for ix := range bookTakers {
		loadBookCh := make(chan []Burnable)
		readyCh := make(chan interface{})

		btParams := &BookTakerParams{
			capacity: 17,
			id:       strconv.Itoa(ix),
			loadCh:   loadBookCh,
			readyCh:  readyCh,
		}

		// Assume that the book taker takes book repeatedly at a specified delay.
		go func() {
			for {
				time.Sleep(1e5)

				go func() {
					readyCh <- true
				}()
			}
		}()

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

	/// When
	for _, taker := range bookTakers {
		go func(taker BookTaker) {
			pileGroup.Supply(taker)
		}(taker)
	}

	time.Sleep(2e9)

	/// Then
	allTakenResult := pileGroup.Taken()
	allTakenMap := make(map[string]int, 0)
	allTakenCount := 0

	for _, result := range allTakenResult {
		takenCount := len(result.bookIds)
		allTakenMap[result.takerID] = allTakenMap[result.takerID] + takenCount
		allTakenCount += takenCount
	}

	if allTakenCount != totalBCount {
		t.Errorf("Should have taken %d, but got %d", totalBCount, allTakenCount)
	}

	for key, value := range allTakenMap {
		t.Logf("Book taker %s took %d books", key, value)

		if value == 0 {
			t.Errorf("%s should have taken some, but took nothing", key)
		}
	}

	loadedLen := len(loadedBooks)

	if loadedLen != totalBCount {
		t.Errorf("Should have taken %d, instead got %d", totalBCount, loadedLen)
	}

	keys := make([]int, 0)

	for key, value := range loadedBooks {
		numericKey, _ := strconv.Atoi(key.UID())
		keys = append(keys, numericKey)

		if value != 1 {
			t.Errorf("%v should have been taken once, but was %d", key, value)
		}
	}

	sort.Ints(keys)
	keyLen := len(keys)

	if keyLen != totalBCount {
		t.Errorf("Should have %d keys, instead got %d", totalBCount, keyLen)
	}
}
