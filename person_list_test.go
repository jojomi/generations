package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonListAddCount(t *testing.T) {
	list := NewPersonList()
	assert.Equal(t, 0, list.Count())

	list.AddPerson(&FlatPerson{ID: "a"})
	list.AddSlice([]Person{&FlatPerson{ID: "b"}, &FlatPerson{ID: "c"}})
	assert.Equal(t, 3, list.Count())

	other := NewPersonList()
	other.AddPerson(&FlatPerson{ID: "d"})
	list.AddList(&other)
	assert.Equal(t, 4, list.Count())
	assert.Equal(t, []string{"a", "b", "c", "d"}, getPersonSliceIDs(&list))
}

func TestPersonListContainsRemove(t *testing.T) {
	a := &FlatPerson{ID: "a"}
	b := &FlatPerson{ID: "b"}

	list := NewPersonList()
	list.AddPerson(a)
	list.AddPerson(b)

	assert.True(t, list.Contains(a))
	assert.False(t, list.Contains(&FlatPerson{ID: "x"}))

	// removing an unknown person is a no-op
	list.Remove(&FlatPerson{ID: "x"})
	assert.Equal(t, 2, list.Count())

	list.Remove(a)
	assert.Equal(t, []string{"b"}, getPersonSliceIDs(&list))
}

func TestPersonListRemoveDuplicatesByBestID(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a"})
	list.AddPerson(&FlatPerson{ID: "b"})
	list.AddPerson(&FlatPerson{ID: "a"})

	list.RemoveDuplicatesByBestID()
	assert.Equal(t, []string{"a", "b"}, getPersonSliceIDs(&list))
}

func TestPersonListExclude(t *testing.T) {
	a := &FlatPerson{ID: "a"}
	b := &FlatPerson{ID: "b"}
	c := &FlatPerson{ID: "c"}

	list := NewPersonList()
	list.AddPerson(a)
	list.AddPerson(b)
	list.AddPerson(c)

	list.Exclude(b)
	assert.Equal(t, []string{"a", "c"}, getPersonSliceIDs(&list))

	filterList := NewPersonList()
	filterList.AddPerson(c)
	list.ExcludeList(&filterList)
	assert.Equal(t, []string{"a"}, getPersonSliceIDs(&list))
}

func TestPersonListDiff(t *testing.T) {
	a := &FlatPerson{ID: "a"}
	b := &FlatPerson{ID: "b"}

	list := NewPersonList()
	list.AddPerson(a)
	list.AddPerson(b)

	other := NewPersonList()
	other.AddPerson(b)

	diff, err := list.Diff(&other)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a"}, getPersonSliceIDs(diff))
}

func TestPersonListHasSameElements(t *testing.T) {
	a := &FlatPerson{ID: "a"}
	b := &FlatPerson{ID: "b"}

	list1 := NewPersonList()
	list1.AddPerson(a)
	list1.AddPerson(b)

	list2 := NewPersonList()
	list2.AddPerson(b)
	list2.AddPerson(a)

	assert.True(t, list1.HasSameElements(&list2))

	list2.AddPerson(&FlatPerson{ID: "c"})
	assert.False(t, list1.HasSameElements(&list2))
}

func TestPersonListSortPersons(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "c"})
	list.AddPerson(&FlatPerson{ID: "a"})
	list.AddPerson(&FlatPerson{ID: "b"})

	list.SortPersons(func(a, b Person) bool {
		return a.GetID() < b.GetID()
	})
	assert.Equal(t, []string{"a", "b", "c"}, getPersonSliceIDs(&list))
}

func TestPersonListGetFiltered(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Gender: "male"})
	list.AddPerson(&FlatPerson{ID: "b", Gender: "female"})

	filtered := list.GetFiltered(func(p Person) bool {
		return p.GetGender() == GenderFemale
	})
	assert.Equal(t, []string{"b"}, getPersonSliceIDs(filtered))
	// original list untouched
	assert.Equal(t, 2, list.Count())
}

func TestPersonListGetAttributes(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a", Attributes: []string{"dead", "farmer"}})
	list.AddPerson(&FlatPerson{ID: "b", Attributes: []string{"farmer"}})

	attrs := list.GetAttributes()
	assert.Len(t, attrs, 2)
	// sorted by attribute name
	assert.Equal(t, "dead", attrs[0].Attribute)
	assert.Equal(t, 1, attrs[0].Persons.Count())
	assert.Equal(t, "farmer", attrs[1].Attribute)
	assert.Equal(t, 2, attrs[1].Persons.Count())
}

func TestGetTreePositionID(t *testing.T) {
	list := NewPersonList()
	list.AddPerson(&FlatPerson{ID: "a"})
	list.AddPerson(&FlatPerson{ID: "b"})

	assert.Equal(t, "-a-b", list.GetTreePositionID())
	assert.Equal(t, "-a-b-c", list.GetTreePositionIDWith(&FlatPerson{ID: "c"}))
}
