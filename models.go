package generations

//go:generate go-enum -f=models.go

type Person interface {
	SetID(id string)
	GetID() string
	GetUUID() string
	GetGender() Gender
	MatchesIDUUID(idUUIDSearches ...string) bool
	MatchesSearch(search string) bool
	GetChildNumber() int
	GetName() Name
	GetBirth() DatePlace
	GetBaptism() DatePlace
	GetDeath() DatePlace
	GetBurial() DatePlace
	// GetChildren returns all children of this person
	GetChildren() ([]Person, error)
	GetMom() (Person, error)
	GetDad() (Person, error)
	// GetPartners returns the list partners that are known for this person
	// A partner is a person that
	// - has been married with this person for any given moment in the past
	// OR
	// - had at least one child with this partner for any given moment in the past
	GetPartners() ([]Person, error)
	// GetChildrenWith returns the list of children of this person with a given partner
	GetChildrenWith(partner Person) ([]Person, error)
	// GetChildrenParents returns the list partners that person has children with (possibly including `nil` iff there is children where no other parent is known)
	GetChildrenParents() ([]Person, error)
	GetAttributes() []string
	AddAttribute(attr string)
	GetImageFilename() string
	GetFloruit() string
	GetJobs() string
	GetComment() string
	IsDummy() bool
}

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

type Name struct {
	First []string `yaml:"first,omitempty"`
	// If the used first name is different from the first element in the .First slice, it can be set using .Used
	Used  string `yaml:"used,omitempty"`
	Last  string `yaml:"last,omitempty"`
	Birth string `yaml:"birth,omitempty"`
	Alias string `yaml:"alias,omitempty"`
	Nick  string `yaml:"nick,omitempty"`
}

type DatePlace struct {
	// need to support things like "before 1932" here
	Date  string `yaml:"date,omitempty"`
	Place string `yaml:"place,omitempty"`
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

func SplitPersons(persons []Person, split Person) (younger []Person, older []Person) {
	isOlder := true
	for _, p := range persons {
		if p.GetID() == split.GetID() {
			isOlder = false
			continue
		}
		if isOlder {
			older = append(older, p)
		} else {
			younger = append(younger, p)
		}
	}
	return
}
