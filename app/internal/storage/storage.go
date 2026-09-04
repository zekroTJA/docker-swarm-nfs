// Package storage persists guestbook entries as individual JSON files inside a
// configurable directory. The file name of an entry is its ID.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/xid"

	"guestbook/internal/guestbook"
)

const entryFileExtension = ".json"

// Storage reads and writes guestbook entries from a directory on disk.
type Storage struct {
	directory string
}

// New creates a Storage rooted at directory. The directory and all missing
// parent directories are created if they do not exist yet.
func New(directory string) (storage *Storage, err error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	return &Storage{directory: directory}, nil
}

// entryFile is the on-disk representation of a guestbook entry. It does not
// contain the ID because the ID is derived from the file name.
type entryFile struct {
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

func (t entryFile) toEntry(id string) guestbook.Entry {
	return guestbook.Entry{
		ID:        id,
		Name:      t.Name,
		Message:   t.Message,
		CreatedAt: t.CreatedAt,
	}
}

// Create stores a new guestbook entry with a freshly generated, time sortable
// ID and returns the persisted entry.
func (t *Storage) Create(name string, message string) (entry guestbook.Entry, err error) {
	id := xid.New().String()
	file := entryFile{
		Name:      name,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}

	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return guestbook.Entry{}, fmt.Errorf("marshal entry: %w", err)
	}

	if err := t.writeFileAtomic(id, raw); err != nil {
		return guestbook.Entry{}, err
	}

	return file.toEntry(id), nil
}

// List returns all stored entries ordered by ascending ID, which is equivalent
// to ascending creation time. If lastID is not empty, only entries with an ID
// greater than lastID are returned.
func (t *Storage) List(lastID string) (entries []guestbook.Entry, err error) {
	dirEntries, err := os.ReadDir(t.directory)
	if err != nil {
		return nil, fmt.Errorf("read storage directory: %w", err)
	}

	entries = make([]guestbook.Entry, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		id, ok := entryID(dirEntry)
		if !ok {
			continue
		}
		if lastID != "" && id <= lastID {
			continue
		}

		entry, err := t.read(id)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i int, j int) bool {
		return entries[i].ID < entries[j].ID
	})

	return entries, nil
}

func (t *Storage) read(id string) (entry guestbook.Entry, err error) {
	path := filepath.Join(t.directory, id+entryFileExtension)
	raw, err := os.ReadFile(path)
	if err != nil {
		return guestbook.Entry{}, fmt.Errorf("read entry file %q: %w", id, err)
	}

	var file entryFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return guestbook.Entry{}, fmt.Errorf("unmarshal entry file %q: %w", id, err)
	}

	return file.toEntry(id), nil
}

// writeFileAtomic writes raw to the entry file of id by writing a temporary
// file first and renaming it, so concurrent readers never observe a partially
// written file.
func (t *Storage) writeFileAtomic(id string, raw []byte) (err error) {
	path := filepath.Join(t.directory, id+entryFileExtension)
	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, raw, 0o644); err != nil {
		return fmt.Errorf("write temporary entry file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("move temporary entry file into place: %w", err)
	}

	return nil
}

// entryID returns the entry ID of a directory entry and whether the directory
// entry is a valid entry file at all.
func entryID(dirEntry os.DirEntry) (id string, ok bool) {
	if dirEntry.IsDir() {
		return "", false
	}

	name := dirEntry.Name()
	if !strings.HasSuffix(name, entryFileExtension) {
		return "", false
	}

	return strings.TrimSuffix(name, entryFileExtension), true
}
