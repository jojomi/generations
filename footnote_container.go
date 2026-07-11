package generations

import (
	"regexp"
)

// labelCleanRegexp strips everything but word characters from labels.
// Attention: might map different inputs to the same label.
var labelCleanRegexp = regexp.MustCompile(`[^\w]+`)

type FootnoteContainer struct {
	footnotes map[string]struct{}
}

func NewFootnoteContainer() *FootnoteContainer {
	return &FootnoteContainer{
		footnotes: make(map[string]struct{}, 0),
	}
}

func (f *FootnoteContainer) Has(search string) bool {
	_, found := f.footnotes[search]
	return found
}

func (f *FootnoteContainer) GetLabel(search string) string {
	_, found := f.footnotes[search]
	if !found {
		return ""
	}
	return f.GetLabelFor(search)
}

func (f *FootnoteContainer) GetLabelFor(input string) string {
	return labelCleanRegexp.ReplaceAllString(input, "")
}

func (f *FootnoteContainer) Add(key string) bool {
	f.footnotes[key] = struct{}{}
	return true
}
