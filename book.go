package goburnbooks

import (
	"math/rand"
	"time"
)

type book struct {
	uid          string
	burnDuration time.Duration
}

func (b *book) String() string {
	return b.uid
}

func (b *book) Burn() {
	time.Sleep(b.burnDuration)
}

// NewBurner returns a burnable book.
func NewBurner(uid string, burnDuration time.Duration) Burner {
	return &book{uid: uid, burnDuration: burnDuration}
}

// NewRandomBurner returns a burnable book with a random burn duration.
func NewRandomBurner(uid string) Burner {
	d := time.Duration(rand.Int()) * time.Microsecond
	return NewBurner(uid, d)
}
