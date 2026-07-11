package gedcom

import (
	"strings"
	"time"
)

// p. 69
type IndividualEventStructure struct {
	Birth BirthEvent
	Death DeathEvent
}

type BirthEvent struct {
	FamcEvent
}

func (b *BirthEvent) StringForLevel(baseLevel int) string {
	return b.FamcEvent.StringForLevel(baseLevel, "BIRT")
}

func (b *BirthEvent) String() string {
	return b.StringForLevel(0)
}

func (b *DeathEvent) StringForLevel(baseLevel int) string {
	return b.FamcEvent.StringForLevel(baseLevel, "DEAT")
}

func (b *DeathEvent) String() string {
	return b.StringForLevel(0)
}

type DeathEvent struct {
	FamcEvent
}

type FamcEvent struct {
	IndividualEventDetail
	Famc string
}

func (b *FamcEvent) IsEmpty() bool {
	return false
}

func (b *FamcEvent) StringForLevel(baseLevel int, tag string) string {
	if b.IsEmpty() {
		return ""
	}
	bl := NewBlock()
	bl.SetLevel(baseLevel)
	if b.Date.Value != "" || b.Place.Name != "" {
		bl.Add(tag)
		if b.Date.Value != "" {
			bl.AddDown("DATE " + b.Date.Value).Up()
		}
		if b.Place.Name != "" {
			bl.AddDown("PLAC " + b.Place.Name).Up()
		}
	}
	return strings.TrimSpace(bl.String())
}

type IndividualEventDetail struct {
	EventDetail
	Age AgeAtEvent
}

type EventDetail struct {
	Type  string
	Date  DateValue
	Place PlaceStructure
}

type PlaceStructure struct {
	Name string
}

type AgeAtEvent struct {
	Value string
}

type DateValue struct {
	Value string
}

func (d *DateValue) String() string {
	return d.Value
}

func (d *DateValue) GetMin() *time.Time {
	// TODO needs more cases (as defined for GEDCOM)
	t, err := time.Parse("02 Jan 2006", strings.Title(d.Value))
	if err != nil {
		return nil
	}
	return &t
}
