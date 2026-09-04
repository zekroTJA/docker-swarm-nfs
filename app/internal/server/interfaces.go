package server

import "guestbook/internal/guestbook"

// Store provides access to the persisted guestbook entries.
type Store interface {
	Create(name string, message string) (entry guestbook.Entry, err error)
	List(lastID string) (entries []guestbook.Entry, err error)
}
