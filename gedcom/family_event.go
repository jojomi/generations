package gedcom

type FamilyEventStructure struct {
	Marriage MarriageFamilyEvent
}

type MarriageFamilyEvent struct {
	FamilyEventDetail
}

type FamilyEventDetail struct {
	Husband string // xref:indi
	Wife    string // xref:indi
	EventDetail
}

func (f *FamilyEventDetail) IsEmpty() bool {
	// TODO clean check
	if f.Date.Value != "" {
		return false
	}
	if f.Place.Name != "" {
		return false
	}
	return true
}

func (f *FamilyEventStructure) String() string {
	b := NewBlock()
	b.SetLevel(0)
	if !f.Marriage.IsEmpty() {
		b.AddDown("MARR") // TODO: MARR Y if no date and place
		if date := f.Marriage.Date.Value; date != "" {
			b.AddDown("DATE " + date).Up()
		}
		if place := f.Marriage.Place.Name; place != "" {
			b.AddDown("PLAC " + place).Up()
		}
	}
	return b.String()
}
