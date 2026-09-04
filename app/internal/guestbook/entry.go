// Package guestbook contains the shared domain types of the guestbook demo
// application.
package guestbook

import "time"

// Entry is a single guestbook entry written by a visitor.
type Entry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}
