package simulator

import "guestbook/internal/guestbook"

// Store persists the synthetic guestbook entries created by the Simulator.
type Store interface {
	Create(name string, message string) (entry guestbook.Entry, err error)
}
