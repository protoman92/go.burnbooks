package goburnbooks

// BookPileGroup represents a group of BookPile.
type BookPileGroup interface {
	BookPile
}

type bookPileGroup struct {
	availablePiles chan BookPile
	bookPiles      []FBookPile
}

func (bpg *bookPileGroup) Supply(taker BookTaker) {
	go func() {
		select {
		case pile := <-bpg.availablePiles:
			pile.Supply(taker)
		}
	}()
}

// Loop supply to store available piles.
func (bpg *bookPileGroup) loopSupply() {
	availablePiles := bpg.availablePiles

	for _, pile := range bpg.bookPiles {
		go func(pile FBookPile) {
			availability := pile.Available()

			for {
				select {
				case <-availability:
					availablePiles <- pile
				}
			}
		}(pile)
	}
}

// NewBookPileGroup creates a new BookPileGroup from a number of BookPiles.
func NewBookPileGroup(piles ...FBookPile) BookPileGroup {
	group := &bookPileGroup{
		availablePiles: make(chan BookPile),
		bookPiles:      piles,
	}

	go group.loopSupply()
	return group
}
