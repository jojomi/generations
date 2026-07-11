package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstLetters(t *testing.T) {
	assert.Equal(t, "", firstLetters("", 4))
	assert.Equal(t, "abc", firstLetters("abc", 4))
	assert.Equal(t, "abcd", firstLetters("abcdef", 4))

	// count must be honored
	assert.Equal(t, "ab", firstLetters("abcdef", 2))

	// rune-safe, not byte-safe
	assert.Equal(t, "Ö", firstLetters("Örjan", 1))
	assert.Equal(t, "Müll", firstLetters("Müller", 4))
}

func TestGetFourFourYearID(t *testing.T) {
	p := &FlatPerson{
		Name:  Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss"},
		Birth: DatePlace{Date: "1777-04-30"},
	}
	id, err := GetFourFourYearID(p)
	assert.Nil(t, err)
	assert.Equal(t, "GausCarl1777", id)

	// used name wins over first name
	p.Name.Used = "Friedrich"
	id, err = GetFourFourYearID(p)
	assert.Nil(t, err)
	assert.Equal(t, "GausFrie1777", id)

	// no birth year
	p = &FlatPerson{Name: Name{First: []string{"Carl"}, Last: "Gauss"}}
	id, err = GetFourFourYearID(p)
	assert.Nil(t, err)
	assert.Equal(t, "GausCarl", id)
}

func TestGetFourFourBirthNameID(t *testing.T) {
	// no birth name: no ID
	id, err := GetFourFourBirthNameID(&FlatPerson{Name: Name{Last: "Gauss"}})
	assert.Nil(t, err)
	assert.Equal(t, "", id)

	id, err = GetFourFourBirthNameID(&FlatPerson{
		Name: Name{First: []string{"Johanna"}, Last: "Gauss", Birth: "Osthoff"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "GausJohaOsth", id)
}

func TestGetLastFirstNameID(t *testing.T) {
	id, err := GetLastFirstNameID(&FlatPerson{
		Name: Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "GaussCarl", id)

	// unknown name
	id, err = GetLastFirstNameID(&FlatPerson{
		Name: Name{First: []string{"?"}, Last: "?"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "", id)
}

func TestAnonymize(t *testing.T) {
	db := NewMemoryDatabase()
	db.Persons = []*FlatPerson{
		{
			Name:  Name{First: []string{"Örjan"}, Last: "Müller"},
			Birth: DatePlace{Date: "1950-04-01", Location: Location{Place: "Musterstadt"}},
			Partners: []FlatRelationship{
				{Marriage: DatePlace{Date: "1975-06-01"}},
			},
		},
		{
			// old enough to be kept as-is
			Name:  Name{First: []string{"Carl"}, Last: "Gauss"},
			Birth: DatePlace{Date: "1777-04-30"},
		},
	}
	db.Anonymize()

	anon := db.Persons[0]
	// initial must be rune-safe (regression: "Ö" was cut mid-rune)
	assert.Equal(t, []string{"Ö."}, anon.Name.First)
	assert.Equal(t, "", anon.Name.Last)
	assert.Equal(t, "1950", anon.Birth.Date)
	assert.Equal(t, "", anon.Birth.Place)
	assert.Equal(t, "1975", anon.Partners[0].Marriage.Date)

	kept := db.Persons[1]
	assert.Equal(t, []string{"Carl"}, kept.Name.First)
	assert.Equal(t, "1777-04-30", kept.Birth.Date)
}
