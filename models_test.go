package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatePlaceEmpty(t *testing.T) {
	assert.True(t, DatePlace{}.Empty())
	assert.False(t, DatePlace{Date: "1990-01-01"}.Empty())
	assert.False(t, DatePlace{Location: Location{Place: "Musterstadt"}}.Empty())
}

func TestGetAgeBegin(t *testing.T) {
	birth := DatePlace{Date: "1970-06-15"}
	marriage := DatePlace{Date: "2000-06-15"}

	assert.Equal(t, 30, marriage.GetAgeBegin(birth))

	// incomplete dates cannot be used
	assert.Equal(t, -1, marriage.GetAgeBegin(DatePlace{}))
	assert.Equal(t, -1, marriage.GetAgeBegin(DatePlace{Date: "1970"}))
	assert.Equal(t, -1, DatePlace{Date: "19??-01-01"}.GetAgeBegin(birth))
	assert.Equal(t, -1, DatePlace{}.GetAgeBegin(birth))
}

func TestGenderIsUnknown(t *testing.T) {
	assert.True(t, GenderUnknown.IsUnknown())
	assert.False(t, GenderMale.IsUnknown())
	assert.False(t, GenderFemale.IsUnknown())
}

func TestSplitPersons(t *testing.T) {
	a := &FlatPerson{ID: "a"}
	b := &FlatPerson{ID: "b"}
	c := &FlatPerson{ID: "c"}

	list := NewPersonList()
	list.AddPerson(a)
	list.AddPerson(b)
	list.AddPerson(c)

	younger, older := SplitPersons(&list, b)
	assert.Equal(t, []string{"a"}, getPersonSliceIDs(&older))
	assert.Equal(t, []string{"c"}, getPersonSliceIDs(&younger))

	// split person not contained: everyone is considered older
	younger, older = SplitPersons(&list, &FlatPerson{ID: "x"})
	assert.Equal(t, []string{"a", "b", "c"}, getPersonSliceIDs(&older))
	assert.Equal(t, 0, younger.Count())
}
