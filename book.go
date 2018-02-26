package goburnbooks

import (
	"time"
)

// BookParams represents the required parameters to set up a Book.
type BookParams struct {
	BurnDuration time.Duration
	UID          string
}

type book struct {
	*BookParams
}

func (b *book) String() string {
	return b.UID
}

func (b *book) Burn() {
	// fmt.Printf("Burning %v\n", b)
	time.Sleep(b.BurnDuration)
}

// NewBook returns a Burnable book.
func NewBook(params *BookParams) Burnable {
	return &book{BookParams: params}
}
