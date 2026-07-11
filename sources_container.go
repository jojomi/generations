package generations

import (
	"strconv"
)

type SourcesContainer struct {
	sources    []string
	sourceMap  map[string]struct{}
	labelBase  string
	mainSource SourceEntry
}

type SourceEntry struct {
	Key  string
	Text string
}

func NewSourcesContainer() *SourcesContainer {
	return &SourcesContainer{
		sources:   make([]string, 0),
		sourceMap: make(map[string]struct{}, 0),
	}
}

func (f *SourcesContainer) Has(search string) bool {
	_, found := f.sourceMap[search]
	return found
}

func (f *SourcesContainer) GetLabel(search string) string {
	if !f.Has(search) {
		return ""
	}
	return f.GetLabelFor(search)
}

func (f *SourcesContainer) SetMainSource(mainSource SourceEntry) {
	f.mainSource = mainSource
}

func (f *SourcesContainer) GetMainSource() SourceEntry {
	return f.mainSource
}

func (f *SourcesContainer) SetLabelBase(base string) {
	f.labelBase = base
}

func (f *SourcesContainer) GetLabelFor(input string) string {
	return f.labelBase + labelCleanRegexp.ReplaceAllString(input, "")
}

// Add returns the key given
func (f *SourcesContainer) Add(key string) string {
	f.sources = append(f.sources, key)
	f.sourceMap[key] = struct{}{}
	return f.GetKeyByIndex(len(f.sources) - 1)
}

func (f *SourcesContainer) GetKey(text string) string {
	for i, source := range f.sources {
		if source != text {
			continue
		}
		return f.GetKeyByIndex(i)
	}
	return ""
}

func (f *SourcesContainer) GetKeyByIndex(i int) string {
	return strconv.Itoa(i + 1) // 1, 2, 3
	//return string(rune(65 + i)) // A, B, C
}

func (f *SourcesContainer) GetNotes() []SourceEntry {
	sources := make([]SourceEntry, 0)
	for i, source := range f.sources {
		sources = append(sources, SourceEntry{
			Key:  f.GetKeyByIndex(i),
			Text: source,
		})
	}
	return sources
}
