package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFootnoteContainer(t *testing.T) {
	f := NewFootnoteContainer()

	assert.False(t, f.Has("note"))
	assert.Equal(t, "", f.GetLabel("note"))

	f.Add("note")
	assert.True(t, f.Has("note"))
	assert.Equal(t, "note", f.GetLabel("note"))

	// labels are stripped of non-word characters
	f.Add("Kirchenbuch, Seite 4")
	assert.Equal(t, "KirchenbuchSeite4", f.GetLabel("Kirchenbuch, Seite 4"))
}

func TestSourcesContainer(t *testing.T) {
	s := NewSourcesContainer()

	assert.False(t, s.Has("Kirchenbuch"))
	assert.Equal(t, "", s.GetLabel("Kirchenbuch"))
	assert.Equal(t, "", s.GetKey("Kirchenbuch"))

	key := s.Add("Kirchenbuch")
	assert.Equal(t, "1", key)
	assert.True(t, s.Has("Kirchenbuch"))
	assert.Equal(t, "1", s.GetKey("Kirchenbuch"))

	key = s.Add("Standesamt")
	assert.Equal(t, "2", key)

	notes := s.GetNotes()
	assert.Len(t, notes, 2)
	assert.Equal(t, SourceEntry{Key: "1", Text: "Kirchenbuch"}, notes[0])
	assert.Equal(t, SourceEntry{Key: "2", Text: "Standesamt"}, notes[1])

	// labels
	s.SetLabelBase("src-")
	assert.Equal(t, "src-Kirchenbuch", s.GetLabel("Kirchenbuch"))

	// main source
	s.SetMainSource(SourceEntry{Key: "1", Text: "Kirchenbuch"})
	assert.Equal(t, "Kirchenbuch", s.GetMainSource().Text)
}

func TestReferencesContainer(t *testing.T) {
	r := NewReferencesContainer()

	refA := Reference{ID: "a", Description: "Kirchenbuch"}
	refB := Reference{ID: "b"}

	assert.False(t, r.Has(refA))
	assert.Equal(t, "", r.GetLabel(refA))
	assert.Equal(t, "", r.GetKey(refA))

	key := r.Add(refA, true)
	assert.Equal(t, "1", key)
	assert.True(t, r.Has(refA))
	assert.Equal(t, "1", r.GetKey(refA))

	key = r.Add(refB, false)
	assert.Equal(t, "2", key)

	assert.Len(t, r.GetSources(), 2)
	main := r.GetMainReferences()
	assert.Len(t, main, 1)
	assert.Equal(t, "a", main[0].Reference.ID)
	nonMain := r.GetNonMainReferences()
	assert.Len(t, nonMain, 1)
	assert.Equal(t, "b", nonMain[0].Reference.ID)

	// labels use the reference string representation
	r.SetLabelBase("ref-")
	assert.Equal(t, "ref-Kirchenbuch", r.GetLabel(refA))
}

func TestLineageContainer(t *testing.T) {
	l := NewLineageContainer()

	assert.Nil(t, l.Get("missing"))

	base := NewPersonList()
	base.AddPerson(&FlatPerson{ID: "root"})
	l.Add("id1", "pos1", base)

	got := l.Get("pos1")
	assert.NotNil(t, got)
	assert.Equal(t, 1, got.Count())

	// extend an existing lineage
	l.AddWithBase("id2", "pos2", "pos1", &FlatPerson{ID: "child"})
	got = l.Get("pos2")
	assert.NotNil(t, got)
	assert.Equal(t, 2, got.Count())
	assert.Equal(t, []string{"root", "child"}, getPersonSliceIDs(got))

	// base does not exist: lineage contains only the person itself
	l.AddWithBase("id3", "pos3", "missing", &FlatPerson{ID: "solo"})
	got = l.Get("pos3")
	assert.NotNil(t, got)
	assert.Equal(t, 1, got.Count())
}
