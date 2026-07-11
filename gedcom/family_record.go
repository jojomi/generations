package gedcom

import (
	"sort"

	"github.com/jojomi/strtpl"
)

type FamilyRecord struct {
	XREF                 string              // xref:fam
	Husband              string              // xref:indi
	Wife                 string              // xref:indi
	Children             []*IndividualRecord // []xref:indi
	FamilyEventStructure *FamilyEventStructure
}

func NewFamilyRecord() *FamilyRecord {
	return &FamilyRecord{
		FamilyEventStructure: &FamilyEventStructure{},
	}
}

func (f *FamilyRecord) AddChild(i *IndividualRecord) *FamilyRecord {
	f.Children = append(f.Children, i)
	return f
}

func (f *FamilyRecord) String() string {
	b := NewBlock()
	b.Add(strtpl.MustEval(`@{{.XREF}}@ FAM`, f))
	if f.Husband != "" {
		b.AddDown(strtpl.MustEval(`HUSB @{{.Husband}}@`, f)).Up()
	}
	if f.Wife != "" {
		b.AddDown(strtpl.MustEval(`WIFE @{{.Wife}}@`, f)).Up()
	}
	for _, child := range f.getSortedChildren() {
		b.AddDown(strtpl.MustEval(`CHILD @{{.XREF}}@`, child)).Up()
	}
	if fam := f.FamilyEventStructure.String(); fam != "" {
		b.AddRaw(fam).Up()
	}
	return b.String()
}

func (f *FamilyRecord) getSortedChildren() []*IndividualRecord {
	sort.SliceStable(f.Children, func(i, j int) bool {
		if f.Children[i].ChildNumber != f.Children[j].ChildNumber {
			return f.Children[i].ChildNumber < f.Children[j].ChildNumber
		}

		// fallback: birth
		a := f.Children[i].Events.Birth.Date.GetMin()
		b := f.Children[j].Events.Birth.Date.GetMin()
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}
		return (*a).Before(*b)
	})
	return f.Children
}

type FamilyRecordSlice []*FamilyRecord

func NewFamilyRecordSlice() *FamilyRecordSlice {
	return &FamilyRecordSlice{}
}
