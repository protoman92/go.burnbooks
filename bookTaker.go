package goburnbooks

import (
	"fmt"
	"time"
)

// BookTaker represents a worker that takes Books for some purposes.
type BookTaker interface {
	Capacity() int
	LoadBooks() chan<- []Burnable

	// Sometimes the BookPile does not have enough Books to fill the taker's load
	// capacity, so the loading process could hang. This timeout allows us to
	// cut the loading prematurely.
	TakeTimeout() time.Duration

	UID() string
}

// BookTakerParams represents all the required parameters to build a book taker.
type BookTakerParams struct {
	capacity  int
	id        string
	loadBooks chan<- []Burnable
	timeout   time.Duration
}

type bookTaker struct {
	*BookTakerParams
}

func (bt *bookTaker) Capacity() int {
	return bt.capacity
}

func (bt *bookTaker) TakeTimeout() time.Duration {
	return bt.timeout
}

func (bt *bookTaker) LoadBooks() chan<- []Burnable {
	return bt.loadBooks
}

func (bt *bookTaker) UID() string {
	return bt.id
}

// NewBookTaker creates a new BookTaker.
func NewBookTaker(params *BookTakerParams) BookTaker {
	return &bookTaker{BookTakerParams: params}
}

// BookTakeResult represents the result of a take operation.
type BookTakeResult struct {
	bookIds []string
	pileID  string
	takerID string
}

func (btr *BookTakeResult) String() string {
	return fmt.Sprintf(
		"Book taker %s took %d books from pile %s",
		btr.takerID,
		len(btr.bookIds),
		btr.pileID,
	)
}
