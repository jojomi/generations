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

type Location struct {
	City    string `yaml:",omitempty"`
	Place   string `yaml:",omitempty"`
	Country string `yaml:",omitempty"`

	CitySources    ReferenceList `yaml:"city_sources,omitempty"`
	PlaceSources   ReferenceList `yaml:"place_sources,omitempty"`
	CountrySources ReferenceList `yaml:"country_sources,omitempty"`
	Sources        ReferenceList `yaml:",omitempty"`
}

type SaintsDay struct {
	SourcedDate `yaml:",inline"`
	Description string `yaml:",omitempty"`
}

func (s SaintsDay) Empty() bool {
	return s.Description == "" && s.SourcedDate.Empty()
}

func (s SaintsDay) String() string {
	return "Saints day @" + s.SourcedDate.Date + " by " + s.Description
}

type SourcedDate struct {
	// need to support things like "before 1932" here
	Date    string        `yaml:",omitempty"`
	Sources ReferenceList `yaml:",omitempty"`
}

func (s SourcedDate) Empty() bool {
	return s.Date == ""
}

type DatePlace struct {
	Location `yaml:",inline,omitempty"`

	// need to support things like "before 1932" here
	Date        string        `yaml:",omitempty"`
	DateSources ReferenceList `yaml:"date_sources,omitempty"`
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

func SplitPersons(personList *PersonList, split Person) (younger PersonList, older PersonList) {
	isOlder := true
	younger = NewPersonList()
	older = NewPersonList()
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
