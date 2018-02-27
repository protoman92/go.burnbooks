package goburnbooks

// BookPileGroup represents a group of BookPile.
type BookPileGroup interface {
	BookPile
	Taken() []*BookTakeResult
}

type bookPileGroup struct {
	bookPiles []FBookPile
	taken     []*BookTakeResult
}

func (bpg *bookPileGroup) Supply(taker BookTaker) {
	for _, pile := range bpg.bookPiles {
		go pile.Supply(taker)
	}
}

func (bpg *bookPileGroup) Taken() []*BookTakeResult {
	return bpg.taken
}

// Loop supply to store available piles and take results.
func (bpg *bookPileGroup) loopSupply() {
	updateTaken := make(chan *BookTakeResult)

	for _, pile := range bpg.bookPiles {
		go func(pile FBookPile) {
			takenResult := pile.TakeResultChannel()

			for {
				select {
				case result := <-takenResult:
					go func() {
						updateTaken <- result
					}()
				}
			}
		}(pile)
	}

	go func() {
		for {
			select {
			case update := <-updateTaken:
				bpg.taken = append(bpg.taken, update)
			}
		}
	}()
}

// NewBookPileGroup creates a new BookPileGroup from a number of BookPiles.
func NewBookPileGroup(piles ...FBookPile) BookPileGroup {
	group := &bookPileGroup{
		bookPiles: piles,
		taken:     make([]*BookTakeResult, 0),
	}

	go group.loopSupply()
	return group
}
