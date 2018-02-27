package goburnbooks

import (
	"time"
)

// BookParams represents the required parameters to set up a Book.
type BookParams struct {
	BurnDuration time.Duration
	ID           string
}

// Book represents a Book.
type Book interface {
	Burnable
}

type book struct {
	*BookParams
}

func (b *book) String() string {
	return b.ID
}

func (b *book) UID() string {
	return b.ID
}

func (b *book) Burn() {
	// fmt.Printf("Burning %v\n", b)
	time.Sleep(b.BurnDuration)
}

// NewBook returns a Burnable and Suppliable book.
func NewBook(params *BookParams) Book {
	return &book{BookParams: params}
}
