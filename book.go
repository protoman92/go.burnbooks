package goburnbooks

import (
	"fmt"
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
	Suppliable
}

type book struct {
	*BookParams
}

func (b *book) String() string {
	return fmt.Sprintf("Book %s", b.ID)
}

func (b *book) BurnableID() string {
	return b.ID
}

func (b *book) SuppliableID() string {
	return b.ID
}

func (b *book) Burn() {
	time.Sleep(b.BurnDuration)
}

// NewBook returns a Burnable and Suppliable book.
func NewBook(params *BookParams) Book {
	return &book{BookParams: params}
}

// ExtractBooksFromBurnables extract Books from a number of Burnables.
func ExtractBooksFromBurnables(burnables ...Burnable) []Book {
	books := make([]Book, 0)

	for _, burnable := range burnables {
		if book, ok := burnable.(Book); ok {
			books = append(books, book)
		}
	}

	return books
}

// ExtractBooksFromSuppliables extract Books from a number of Suppliables.
func ExtractBooksFromSuppliables(suppliables ...Suppliable) []Book {
	books := make([]Book, 0)

	for _, suppliable := range suppliables {
		if book, ok := suppliable.(Book); ok {
			books = append(books, book)
		}
	}

	return books
}

// ExtractBurnablesFromSuppliables extract Burnables from a number of Suppliables.
func ExtractBurnablesFromSuppliables(suppliables ...Suppliable) []Burnable {
	burnables := make([]Burnable, 0)

	for _, suppliable := range suppliables {
		if book, ok := suppliable.(Burnable); ok {
			burnables = append(burnables, book)
		}
	}

	return burnables
}
