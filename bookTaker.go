package goburnbooks

import (
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
}

// BookTakerParams represents all the required parameters to build a book taker.
type BookTakerParams struct {
	capacity  int
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

// NewBookTaker creates a new BookTaker.
func NewBookTaker(params *BookTakerParams) BookTaker {
	return &bookTaker{BookTakerParams: params}
}
