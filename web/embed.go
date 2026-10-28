// Package web embeds templates and static assets into the binary
// (SPEC §1 single-binary goal). The generated app.css is a build
// artifact produced by the Tailwind CLI; it is never hand-edited.
package web

import "embed"

//go:embed templates static
var FS embed.FS
