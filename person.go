package generations

import "time"

// Person interface
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
	// GetDeathAge returns the age in years at the given point of time. -1 iff the age can't be determined.
	GetAge(now time.Time) int
	// GetDeathAge returns the age in years when the person died. -1 iff the age can't be determined.
	GetDeathAge() int
	GetBurial() DatePlace
	// GetChildren returns all children of this person
	GetChildren() (PersonList, error)
	GetMom() (Person, error)
	GetDad() (Person, error)
	GetRelationships() []Relationship
	// GetPartners returns the list partners that are known for this person
	// A partner is a person that
	// - has been married with this person for any given moment in the past
	// OR
	// - had at least one child with this partner for any given moment in the past
	GetPartners() (PersonList, error)
	// GetChildrenWith returns the list of children of this person with a given partner
	GetChildrenWith(partner Person) (PersonList, error)
	// GetChildrenParents returns the list partners that person has children with (possibly including `nil` iff there is children where no other parent is known)
	GetChildrenParents() (PersonList, error)
	GetAttributes() []string
	AddAttribute(attr string)
	GetImageFilename() string
	SetImageFilename(filename string)
	GetFloruit() string
	GetJobs() string
	SetJobs(jobs string)
	GetComment() string
	SetComment(comment string)
	IsDummy() bool
}
