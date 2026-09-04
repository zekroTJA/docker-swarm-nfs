// Package web provides the embedded static assets of the guestbook web page.
package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:assets
var embedded embed.FS

// Assets is the file system holding the static web page assets (index.html,
// app.js and style.css).
var Assets = mustSub(embedded, "assets")

func mustSub(source fs.FS, directory string) fs.FS {
	sub, err := fs.Sub(source, directory)
	if err != nil {
		panic(fmt.Sprintf("web: build sub filesystem %q: %v", directory, err))
	}

	return sub
}
