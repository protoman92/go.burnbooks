package goburnbooks

import (
	"fmt"
)

// BookTaker represents a worker that takes Books for some purposes.
type BookTaker interface {
	Capacity() int
	LoadChannel() chan<- []Burnable
	ReadyChannel() <-chan interface{}
	UID() string
}

// BookTakerParams represents all the required parameters to build a book taker.
type BookTakerParams struct {
	capacity int
	id       string
	loadCh   chan<- []Burnable
	readyCh  chan interface{}
}

type bookTaker struct {
	*BookTakerParams
}

func (bt *bookTaker) Capacity() int {
	return bt.capacity
}

func (bt *bookTaker) LoadChannel() chan<- []Burnable {
	return bt.loadCh
}

func (bt *bookTaker) ReadyChannel() <-chan interface{} {
	return bt.readyCh
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
