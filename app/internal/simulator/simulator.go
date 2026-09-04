// Package simulator periodically creates synthetic guestbook entries to
// simulate visitor activity for demonstration purposes.
package simulator

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"
)

//go:embed data/usernames.txt data/messages.txt
var sampleData embed.FS

// Simulator writes a randomly composed guestbook entry to the store on a fixed
// interval.
type Simulator struct {
	store    Store
	interval time.Duration

	usernames []string
	messages  []string
	random    *rand.Rand
}

// New creates a Simulator that writes one synthetic entry to store every
// interval. The pools of names and messages are loaded from embedded text
// files.
func New(store Store, interval time.Duration) (simulator *Simulator, err error) {
	usernames, err := readLines("data/usernames.txt")
	if err != nil {
		return nil, err
	}

	messages, err := readLines("data/messages.txt")
	if err != nil {
		return nil, err
	}

	return &Simulator{
		store:     store,
		interval:  interval,
		usernames: usernames,
		messages:  messages,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Run creates synthetic entries until ctx is cancelled.
func (t *Simulator) Run(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			name := t.usernames[t.random.Intn(len(t.usernames))]
			message := t.messages[t.random.Intn(len(t.messages))]
			if _, err := t.store.Create(name, message); err != nil {
				slog.Error("failed creating synthetic guestbook entry", "error", err)
			}
		}
	}
}

func readLines(name string) (lines []string, err error) {
	raw, err := sampleData.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read embedded sample file %q: %w", name, err)
	}

	lines = make([]string, 0)
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("embedded sample file %q is empty", name)
	}

	return lines, nil
}
