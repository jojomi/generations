package generations

import (
	"strconv"
)

type ReferencesContainer struct {
	sources   []ReferenceEntry
	sourceMap map[string]struct{}
	labelBase string
}

type ReferenceEntry struct {
	Key          string
	Reference    Reference
	IsMainSource bool
}

func NewReferencesContainer() *ReferencesContainer {
	return &ReferencesContainer{
		sources:   make([]ReferenceEntry, 0),
		sourceMap: make(map[string]struct{}, 0),
	}
}

func (f *ReferencesContainer) Has(search Reference) bool {
	_, found := f.sourceMap[search.String()]
	return found
}

func (f *ReferencesContainer) GetLabel(search Reference) string {
	if !f.Has(search) {
		return ""
	}
	return f.GetLabelFor(search)
}

func (f *ReferencesContainer) SetLabelBase(base string) {
	f.labelBase = base
}

func (f *ReferencesContainer) GetLabelFor(source Reference) string {
	return f.labelBase + labelCleanRegexp.ReplaceAllString(source.String(), "")
}

// Add returns the key given
func (f *ReferencesContainer) Add(source Reference, isMainSource bool) string {
	// NOOP
	/// return ""

	sourceEntry := ReferenceEntry{
		Key:          f.GetKeyByIndex(len(f.sources)),
		Reference:    source,
		IsMainSource: isMainSource,
	}
	f.sources = append(f.sources, sourceEntry)
	f.sourceMap[sourceEntry.Reference.String()] = struct{}{}
	return f.GetKeyByIndex(len(f.sources) - 1)
}

func (f *ReferencesContainer) GetKey(source Reference) string {
	// NOOP
	/// return ""

	for i, s := range f.sources {
		if s.Reference.String() != source.String() {
			continue
		}
		return f.GetKeyByIndex(i)
	}
	return ""
}

func (f *ReferencesContainer) GetKeyByIndex(i int) string {
	return strconv.Itoa(i + 1) // 1, 2, 3
	//return string(rune(65 + i)) // A, B, C
}

func (f *ReferencesContainer) GetSources() []ReferenceEntry {
	return f.sources
}

func (f *ReferencesContainer) GetMainReferences() []ReferenceEntry {
	result := make([]ReferenceEntry, 0, 2)
	for _, s := range f.sources {
		if !s.IsMainSource {
			continue
		}
		result = append(result, s)
	}
	return result
}

func (f *ReferencesContainer) GetNonMainReferences() []ReferenceEntry {
	result := make([]ReferenceEntry, 0, len(f.sources))
	for _, s := range f.sources {
		if s.IsMainSource {
			continue
		}
		result = append(result, s)
	}
	return result
}
