package goburnbooks

// Available represents something that can be marked as available.
type Available interface {
	Available() <-chan bool
}
