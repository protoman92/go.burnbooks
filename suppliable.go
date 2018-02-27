package goburnbooks

// Suppliable represents something that can be taken from a SupplyPile.
type Suppliable interface {
	UID() string
}
