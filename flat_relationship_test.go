package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeCouple(marriageDate string) (*MemoryDatabase, *FlatPerson, *FlatPerson) {
	db := NewMemoryDatabase()
	a := &FlatPerson{
		ID:       "a",
		Gender:   "female",
		Database: db,
	}
	b := &FlatPerson{
		ID:       "b",
		Gender:   "male",
		Database: db,
	}
	a.Partners = []FlatRelationship{
		{PartnerID: "b", Marriage: DatePlace{Date: marriageDate}},
	}
	db.Persons = []*FlatPerson{a, b}
	return db, a, b
}

func TestSetChildrenCount(t *testing.T) {
	r := &FlatRelationship{}
	assert.False(t, r.HasChildrenCount())

	count := uint(3)
	r.SetChildrenCount(&count)
	assert.True(t, r.HasChildrenCount())
	assert.Equal(t, uint(3), *r.GetChildrenCount())

	// via the Relationship interface as returned by GetRelationships
	_, a, _ := makeCouple("1950-06-01")
	rel := a.GetRelationships()[0]
	rel.SetChildrenCount(&count)
	assert.True(t, rel.HasChildrenCount())
	assert.Equal(t, uint(3), *rel.GetChildrenCount())
}

func TestGetMarriageDuration(t *testing.T) {
	ref := testDate("2000-06-01")

	// both partners alive: duration until reference date, not finished
	_, a, _ := makeCouple("1950-06-01")
	d := a.GetRelationships()[0].GetMarriageDuration(a, ref)
	assert.Nil(t, d.Error)
	assert.Equal(t, 50, d.Years)
	assert.False(t, d.IsFinished)

	// ended by death of one partner
	_, a, b := makeCouple("1950-06-01")
	b.Death = DatePlace{Date: "1980-03-01"}
	d = a.GetRelationships()[0].GetMarriageDuration(a, ref)
	assert.Nil(t, d.Error)
	assert.Equal(t, 29, d.Years)
	assert.True(t, d.IsFinished)

	// no marriage date: no duration (must not be negative)
	_, a, _ = makeCouple("")
	d = a.GetRelationships()[0].GetMarriageDuration(a, ref)
	assert.NotNil(t, d.Error)

	// partner dead without death date: no duration (must not be negative)
	_, a, b = makeCouple("1950-06-01")
	b.Dead = boolPtr(true)
	d = a.GetRelationships()[0].GetMarriageDuration(a, ref)
	assert.NotNil(t, d.Error)
}

func TestGetPartner(t *testing.T) {
	_, a, b := makeCouple("1950-06-01")

	partner, err := a.GetRelationships()[0].GetPartner()
	assert.Nil(t, err)
	assert.Equal(t, b.GetID(), partner.GetID())

	// unknown partner id
	a.Partners[0].PartnerID = "unknown"
	_, err = a.GetRelationships()[0].GetPartner()
	assert.NotNil(t, err)
	assert.True(t, a.GetRelationships()[0].GetPartnerOrDummy().IsDummy())
}
