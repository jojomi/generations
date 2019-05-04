package generations

import (
	"time"

	age "github.com/bearbin/go-age"
)

//go:generate go-enum -f=models.go

// Gender x ENUM(
// unknown
// male
// female
// )
type Gender int8

type Marriage struct {
	DatePlace
	Partner Person
}

type DatePlace struct {
	// need to support things like "before 1932" here
	Date  string `yaml:"date,omitempty"`
	Place string `yaml:"place,omitempty"`
}

func (d DatePlace) GetAgeBegin(other DatePlace) int {
	if other.Empty() || len(other.Date) != 10 {
		return -1
	}
	otherTime, err := time.Parse("2006-01-02", other.Date)
	if err != nil {
		return -1
	}

	if d.Empty() || len(d.Date) != 10 {
		return -1
	}
	dateTime, err := time.Parse("2006-01-02", d.Date)
	if err != nil {
		return -1
	}

	return age.AgeAt(otherTime, dateTime)
}

func (g Gender) IsUnknown() bool {
	return g == GenderUnknown
}

func (d DatePlace) Empty() bool {
	return d.Date == "" && d.Place == ""
}

func (n Name) Empty() bool {
	return len(n.First) == 0 && n.Last == "" && n.Birth == ""
}

func SplitPersons(personList PersonList, split Person) (younger PersonList, older PersonList) {
	isOlder := true
	younger = NewPersonList(nil)
	older = NewPersonList(nil)
	for _, p := range personList.GetPersons() {
		if p.GetID() == split.GetID() {
			isOlder = false
			continue
		}
		if isOlder {
			older.AddPerson(p)
		} else {
			younger.AddPerson(p)
		}
	}
	return
}
