package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNameLength(t *testing.T) {
	assert.Equal(t, 0, getNameLength(""))
	assert.Equal(t, 4, getNameLength("Carl"))
	// separators and unknowns are not counted
	assert.Equal(t, 8, getNameLength("Anna-Lena"))
	assert.Equal(t, 8, getNameLength("Anna Le na"))
	assert.Equal(t, 0, getNameLength("??"))
	// umlauts count as one letter
	assert.Equal(t, 6, getNameLength("Müller"))
}

func TestGenderCounts(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Gender: "male"})
	list.AddPerson(&FlatPerson{ID: "b", Gender: "female"})
	list.AddPerson(&FlatPerson{ID: "c", Gender: "female"})
	list.AddPerson(&FlatPerson{ID: "d"})

	assert.Equal(t, 1, s.MaleCount(&list))
	assert.Equal(t, 2, s.FemaleCount(&list))
}

func TestAliveDeadCounts(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Birth: DatePlace{Date: "1990-01-01"}})
	list.AddPerson(&FlatPerson{ID: "b", Birth: DatePlace{Date: "1900-01-01"}, Death: DatePlace{Date: "1980-01-01"}})

	assert.Equal(t, 1, s.AliveCount(&list))
	assert.Equal(t, 1, s.DeadCount(&list))
}

func TestTotalAgeLiving(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Birth: DatePlace{Date: "1990-01-01"}})
	list.AddPerson(&FlatPerson{ID: "b", Birth: DatePlace{Date: "2000-01-01"}})
	// no birth date: cannot contribute (must not panic and must not distort the sum)
	list.AddPerson(&FlatPerson{ID: "c"})
	// dead persons don't count
	list.AddPerson(&FlatPerson{ID: "d", Birth: DatePlace{Date: "1900-01-01"}, Death: DatePlace{Date: "1980-01-01"}})

	assert.Equal(t, int64(30+20), s.TotalAgeLiving(&list))
}

func TestOldestAndNewestBirths(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Birth: DatePlace{Date: "1950-01-01"}})
	list.AddPerson(&FlatPerson{ID: "b", Birth: DatePlace{Date: "1970-01-01"}})
	list.AddPerson(&FlatPerson{ID: "c", Birth: DatePlace{Date: "1990-01-01"}})
	// no birth date: not ranked
	list.AddPerson(&FlatPerson{ID: "d"})

	oldest := s.OldestBirths(&list)
	assert.Equal(t, 3, oldest.Len())
	assert.Equal(t, []string{"a"}, getPersonSliceIDs(oldest.GetItems()[0].Persons))

	newest := s.NewestBirths(&list)
	assert.Equal(t, 3, newest.Len())
	assert.Equal(t, []string{"c"}, getPersonSliceIDs(newest.GetItems()[0].Persons))

	// GetTop caps the result
	assert.Len(t, oldest.GetTop(2), 2)
	assert.Len(t, oldest.GetTop(10), 3)
}

func TestMarriageDays(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	db := NewMemoryDatabase()
	a := &FlatPerson{ID: "a", Gender: "female", Database: db}
	b := &FlatPerson{ID: "b", Gender: "male", Database: db}
	c := &FlatPerson{ID: "c", Gender: "male", Database: db}
	// one relationship without marriage date must not hide the other one (regression)
	a.Partners = []FlatRelationship{
		{PartnerID: "c"},
		{PartnerID: "b", Marriage: DatePlace{Date: "1950-06-01"}},
	}
	db.Persons = []*FlatPerson{a, b, c}

	list := NewPersonList()
	list.AddPerson(a)

	result := s.MarriageDays(&list)
	assert.Equal(t, 1, result.Len())
	assert.Equal(t, []string{"a", "b"}, getPersonSliceIDs(result.GetItems()[0].Persons))
}

func TestAvgMarriageAge(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	// empty list must not return NaN
	empty := NewPersonList()
	assert.Equal(t, 0.0, s.AvgMarriageAge(&empty))

	db := NewMemoryDatabase()
	a := &FlatPerson{
		ID:       "a",
		Gender:   "female",
		Birth:    DatePlace{Date: "1950-06-01"},
		Database: db,
	}
	a.Partners = []FlatRelationship{
		{PartnerID: "b", Marriage: DatePlace{Date: "1975-06-01"}},
	}
	b := &FlatPerson{ID: "b", Gender: "male", Database: db}
	db.Persons = []*FlatPerson{a, b}

	list := NewPersonList()
	list.AddPerson(a)

	assert.Equal(t, 25.0, s.AvgMarriageAge(&list))
}

func TestHistogramHelpers(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Gender: "male"})
	list.AddPerson(&FlatPerson{ID: "b", Gender: "male"})
	list.AddPerson(&FlatPerson{ID: "c", Gender: "female"})
	list.AddPerson(&FlatPerson{ID: "d"})

	histogram := getHistogram(&list, func(p Person) string {
		if p.GetGender().IsUnknown() {
			return "" // empty keys are skipped
		}
		return p.GetGender().String()
	})
	assert.Len(t, histogram, 2)
	assert.Equal(t, 2, histogram["male"].Count())
	assert.Equal(t, 1, histogram["female"].Count())

	unique := getHistogramUnique(histogram)
	assert.Len(t, unique, 1)
	assert.NotNil(t, unique["female"])

	nonUnique := getHistogramNonUnique(histogram)
	assert.Len(t, nonUnique, 1)
	assert.NotNil(t, nonUnique["male"])
}

func TestMostCommonFirstNames(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Name: Name{First: []string{"Carl"}}})
	list.AddPerson(&FlatPerson{ID: "b", Name: Name{First: []string{"Carl"}}})
	list.AddPerson(&FlatPerson{ID: "c", Name: Name{First: []string{"Anna"}}})

	result := s.MostCommonFirstNames(&list)
	assert.Equal(t, 2, result.Len())
	top := result.GetItems()[0]
	assert.Equal(t, "Carl", top.Value)
	assert.Equal(t, 2, top.RankingValue)
}

func TestMultipleAncestors(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	// Case 1: Person appears twice as ancestor (depth 2 and 3) -> included
	a := &FlatPerson{ID: "a"}
	a.AddTreePositionID("-proband-a")
	a.AddTreePositionID("-proband-x-a")

	// Case 2: Person appears only once -> excluded
	b := &FlatPerson{ID: "b"}
	b.AddTreePositionID("-proband-b")

	// Case 3: Person appears 3 times, 2 as ancestor + 1 as proband -> included, count=3
	c := &FlatPerson{ID: "c"}
	c.AddTreePositionID("-proband-c")
	c.AddTreePositionID("-proband-y-c")
	c.AddTreePositionID("-proband-y-z-c")

	// Case 4: Person appears twice but only as proband/extra (depth 1) -> excluded
	d := &FlatPerson{ID: "d"}
	d.AddTreePositionID("-d")
	d.AddTreePositionID("-x")

	list := NewPersonList()
	list.AddPerson(a)
	list.AddPerson(b)
	list.AddPerson(c)
	list.AddPerson(d)

	result := s.MultipleAncestors(&list)
	assert.Equal(t, 2, result.Len())

	// Sorted by count descending: c (3) first, then a (2)
	assert.Equal(t, []string{"c"}, getPersonSliceIDs(result.GetItems()[0].Persons))
	assert.Equal(t, 3, result.GetItems()[0].RankingValue)
	assert.Equal(t, []string{"a"}, getPersonSliceIDs(result.GetItems()[1].Persons))
	assert.Equal(t, 2, result.GetItems()[1].RankingValue)
}

func TestMultipleAncestorsEmpty(t *testing.T) {
	s := NewStats(testDate("2020-06-15"))

	// Only single-occurrence persons
	a := &FlatPerson{ID: "a"}
	a.AddTreePositionID("-proband-a")

	list := NewPersonList()
	list.AddPerson(a)

	result := s.MultipleAncestors(&list)
	assert.Equal(t, 0, result.Len())
}
