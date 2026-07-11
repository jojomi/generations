package generations

import "time"

// Person interface
type Person interface {
	Equals(p2 Person) bool

	SetID(id string)
	GetID() string
	GetUUID() string
	GetBestID() string
	GetGender() Gender
	MatchesIDUUID(idUUIDSearches ...string) bool
	MatchesSearch(search string) bool
	GetChildNumber() *uint
	SetChildNumber(childNumber *uint)
	GetRelationsCount() *uint
	SetRelationsCount(relationsCount *uint)
	GetName() Name
	GetBirth() DatePlace
	GetBaptism() DatePlace
	GetSaintsDay() SaintsDay
	GetDeath() DatePlace
	GetDeadState() *bool
	// GetAge returns the age in years at the given point of time. -1 iff the age can't be determined.
	GetAge(now time.Time) int
	// GetDeathAge returns the age in years when the person died. -1 iff the age can't be determined.
	GetDeathAge() int
	GetBurial() DatePlace
	// GetChildren returns all children of this person
	GetChildren() (*PersonList, error)
	HasChildrenCount() bool
	GetChildrenCount() int
	GetRawMom() string
	GetMom() (Person, error)
	SetRawMom(string)
	GetRawDad() string
	GetDad() (Person, error)
	SetRawDad(string)
	GetSiblings() (*SiblingList, error)
	GetRelationships() []Relationship
	GetRelationshipWith(partner Person) Relationship
	// GetPartners returns the list partners that are known for this person
	// A partner is a person that
	// - has been married with this person for any given moment in the past
	// OR
	// - had at least one child with this partner for any given moment in the past
	GetPartners() (*PersonList, error)
	// GetChildrenWith returns the list of children of this person with a given partner
	GetChildrenWith(partner Person) (*PersonList, error)
	// GetChildrenParents returns the list partners that person has children with (possibly including `nil` iff there is children where no other parent is known)
	GetChildrenParents() (*PersonList, error)
	GetAttributes() []string
	AddAttribute(attr string)
	GetImageFilename() string
	SetImageFilename(filename string)
	GetResidences() []Residence
	AddResidence(residence Residence)
	SetResidences(residences []Residence)
	GetJobs() []Job
	SetJobs(jobs []Job)
	AddJob(job Job)
	GetBiographyElements() []BiographyElement
	AddBiographyElement(biographyElement BiographyElement)
	SetBiographyElements(biographyElements []BiographyElement)
	GetComments() []Comment
	AddComment(comment Comment)
	SetComments(comments []Comment)
	SetStopPerson(value bool)
	IsStopPerson() bool

	IsDummy() bool
	IsEmpty() bool
	IsAlive(referenceTime time.Time) bool
	IsSpitzahn() bool

	GetTreePositionIDs() []string
	GetMainTreePositionID() string
	AddTreePositionID(ID string) Person
	ClearTreePositionIDs() Person

	// source management
	GetSources() ReferenceList
	GetGenderSources() ReferenceList
	/*GetChildNumberSources() ReferenceList
	GetRelationsCountSources() ReferenceList*/
	/*GetDeadStateSources() ReferenceList
	GetAgeSources() ReferenceList
	GetDeathAgeSources() ReferenceList */
	/* GetChildrenCountSources() ReferenceList
	GetMomSources() ReferenceList
	GetDadSources() ReferenceList
	GetSiblingsSources() ReferenceList
	GetRelationshipsSources() ReferenceList
	GetAttributesSources() ReferenceList
	GetImageFilenameSources() ReferenceList */

	Copy() Person
	String() string
}
