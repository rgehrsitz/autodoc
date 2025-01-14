// autodoc/pkg/wiki/helpers.go

package wiki

import (
	"github.com/russross/blackfriday/v2"
)

// NavItem defines a single navigation entry
type NavItem struct {
	Title    string
	URL      string
	Active   bool
	Children []NavItem
}

// renderMarkdown converts markdown content to HTML
func renderMarkdown(content string) string {
	md := blackfriday.Run([]byte(content),
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.AutoHeadingIDs|
				blackfriday.NoEmptyLineBeforeBlock,
		),
	)
	return string(md)
}
