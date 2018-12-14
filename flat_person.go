package generations

import (
	"sort"
	"time"

	age "github.com/bearbin/go-age"
	"github.com/jojomi/strtpl"
)

type FlatPerson struct {
	Dummy         bool               `yaml:"-"`
	ID            string             `yaml:"id,omitempty"`
	UUID          string             `yaml:"uuid,omitempty"`
	ChildNumber   int                `yaml:"child-number,omitempty"`
	Name          Name               `yaml:"name,omitempty"`
	Gender        string             `yaml:"gender,omitempty"`
	Birth         DatePlace          `yaml:"birth,omitempty"`
	Baptism       DatePlace          `yaml:"baptism,omitempty"`
	Death         DatePlace          `yaml:"death,omitempty"`
	Burial        DatePlace          `yaml:"burial,omitempty"`
	Mom           string             `yaml:"mom,omitempty"`
	Dad           string             `yaml:"dad,omitempty"`
	Partners      []FlatRelationship `yaml:"partners,omitempty"`
	Attributes    []string           `yaml:"attributes,omitempty"`
	ImageFilename string             `yaml:"image,omitempty"`
	Floruit       string             `yaml:"floruit,omitempty"`
	Jobs          string             `yaml:"jobs,omitempty"`
	Comment       string             `yaml:"comment,omitempty"`

	Database *MemoryDatabase `yaml:"-"`
}

type FlatRelationship struct {
	PartnerID  string    `yaml:"partner_id,omitempty"`
	Engagement DatePlace `yaml:"engagement,omitempty"`
	Marriage   DatePlace `yaml:"marriage,omitempty"`
	Divorce    DatePlace `yaml:"divorce,omitempty"`
}

func NewDummyFlatPerson() *FlatPerson {
	return &FlatPerson{
		ID:    "<dummy>",
		Dummy: true,
	}
}

func (d *FlatPerson) SetID(id string) {
	d.ID = id
}

func (d *FlatPerson) GetID() string {
	return d.ID
}

func (d *FlatPerson) GetUUID() string {
	return d.UUID
}

func (d *FlatPerson) MatchesIDUUID(idUUIDSearches ...string) bool {
	for _, search := range idUUIDSearches {
		if search == "" {
			continue
		}
		if d.GetUUID() == search {
			return true
		}
		if d.GetID() == search {
			return true
		}
	}
	return false
}

func (d *FlatPerson) MatchesSearch(search string) bool {
	if d.MatchesIDUUID(search) {
		return true
	}
	name := d.GetName()
	if len(name.First) > 0 {
		if name.First[0]+" "+name.Last == search {
			return true
		}
	}
	return false
}

func (d *FlatPerson) GetGender() Gender {
	gender, err := ParseGender(d.Gender)
	if err != nil {
		return GenderUnknown
	}
	return gender
}

func (d *FlatPerson) GetName() Name {
	return d.Name
}

func (d *FlatPerson) GetChildNumber() int {
	return d.ChildNumber
}

func (d *FlatPerson) GetBirth() DatePlace {
	return d.Birth
}

func (d *FlatPerson) GetBaptism() DatePlace {
	return d.Baptism
}

func (d *FlatPerson) GetDeath() DatePlace {
	return d.Death
}

func (d *FlatPerson) GetAge(now time.Time) int {
	birth := d.GetBirth()

	if now.IsZero() {
		return -1
	}

	if birth.Empty() || len(birth.Date) != 10 {
		return -1
	}
	birthTime, err := time.Parse("2006-01-02", birth.Date)
	if err != nil {
		return -1
	}

	return age.AgeAt(birthTime, now)
}

func (d *FlatPerson) GetDeathAge() int {
	birth := d.GetBirth()
	death := d.GetDeath()

	if birth.Empty() || len(birth.Date) != 10 {
		return -1
	}
	birthTime, err := time.Parse("2006-01-02", birth.Date)
	if err != nil {
		return -1
	}

	if death.Empty() || len(death.Date) != 10 {
		return -1
	}
	deathTime, err := time.Parse("2006-01-02", death.Date)
	if err != nil {
		return -1
	}

	return age.AgeAt(birthTime, deathTime)
}

func (d *FlatPerson) GetBurial() DatePlace {
	return d.Burial
}

func (d *FlatPerson) GetMom() (Person, error) {
	if d.Mom == "" {
		return NewDummyFlatPerson(), nil
	}
	mom, err := d.Database.GetByID(d.Mom)
	if err != nil {
		return NewDummyFlatPerson(), err
	}
	return mom, nil
}

func (d *FlatPerson) GetDad() (Person, error) {
	if d.Dad == "" {
		return NewDummyFlatPerson(), nil
	}
	dad, err := d.Database.GetByID(d.Dad)
	if err != nil {
		return NewDummyFlatPerson(), err
	}
	return dad, nil
}

func (d *FlatPerson) GetPartners() ([]Person, error) {
	result := []Person{}

	// find explicit partners
	for _, partner := range d.Partners {
		for _, person := range d.Database.Persons {
			if person.MatchesIDUUID(partner.PartnerID) {
				result = append(result, person)
			}
		}
	}

	// find partners through common children
	childrenPartners, err := d.GetChildrenParents()
	if err != nil {
		return []Person{}, err
	}

	// merge results
	intermediateResult := deduplicatePersonSlices(mergePersonSlices(childrenPartners, result))

	finalResult := make([]Person, len(intermediateResult))
	for i, p := range intermediateResult {
		finalResult[len(intermediateResult)-1-i] = p
	}
	return finalResult, nil
}

func (d *FlatPerson) GetChildrenParents() ([]Person, error) {
	result := []Person{}
	parentsSeen := make(map[string]struct{}, 0)
	var (
		mom       Person
		dad       Person
		candidate Person
		ok        bool
		err       error
	)
	children, err := d.GetChildren()
	if err != nil {
		return result, err
	}
	for _, child := range children {
		mom, err = child.GetMom()
		if err != nil {
			return result, err
		}
		dad, err = child.GetDad()
		if err != nil {
			return result, err
		}

		// if this FlatPerson is mom or dad the other one is a candidate parent to be returned
		if mom != nil && mom.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			candidate = dad
		}
		if dad != nil && dad.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			candidate = mom
		}

		// if the candidate has already been seen, don't add it to the parent list again
		if candidate == nil {
			candidate = NewDummyFlatPerson()
		}

		if _, ok = parentsSeen[candidate.GetID()]; ok {
			continue
		}
		parentsSeen[candidate.GetID()] = struct{}{}

		result = append(result, candidate)
	}
	return result, nil
}

func (d *FlatPerson) GetChildren() ([]Person, error) {
	result := []Person{}

	var (
		mom Person
		dad Person
		err error
	)
	for _, child := range d.Database.Persons {
		mom, err = child.GetMom()
		if err != nil {
			return []Person{}, err
		}
		if mom != nil && mom.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			result = append(result, child)
			continue
		}
		dad, err = child.GetDad()
		if err != nil {
			return []Person{}, err
		}
		if dad != nil && dad.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			result = append(result, child)
			continue
		}
	}

	return result, nil
}

func (d *FlatPerson) GetChildrenWith(partner Person) ([]Person, error) {
	result := []Person{}
	children, err := d.GetChildren()
	if err != nil {
		return result, err
	}
	for _, child := range children {
		mom, err := child.GetMom()
		if err != nil {
			return []Person{}, err
		}
		dad, err := child.GetDad()
		if err != nil {
			return []Person{}, err
		}
		otherParent := getOtherPerson(mom, dad, d)
		if partner.IsDummy() {
			if otherParent.IsDummy() {
				result = append(result, child)
			}
			continue
		}

		if otherParent.IsDummy() {
			continue
		}
		if otherParent.GetID() == partner.GetID() {
			result = append(result, child)
		}
	}

	// sort children: first by child-number, then by date of birth
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].GetChildNumber() != result[j].GetChildNumber() {
			return result[i].GetChildNumber() < result[j].GetChildNumber()
		}
		return result[i].GetBirth().Date < result[j].GetBirth().Date
	})

	return result, nil
}

func (d *FlatPerson) GetAttributes() []string {
	result := d.Attributes

	// feature: auto-detect dead person and add "dead" attribute automagically
	isDead := !d.GetDeath().Empty() || !d.GetBurial().Empty()
	if isDead {
		hasDeadAttribute := false
		for _, a := range d.Attributes {
			if a == "dead" {
				hasDeadAttribute = true
				break
			}
		}
		if !hasDeadAttribute {
			result = append(result, "dead")
		}
	}
	return result
}

func (d *FlatPerson) AddAttribute(attr string) {
	d.Attributes = append(d.Attributes, attr)
}

func (d *FlatPerson) GetImageFilename() string {
	return d.ImageFilename
}

func (d *FlatPerson) GetJobs() string {
	return d.Jobs
}

func (d *FlatPerson) GetFloruit() string {
	return d.Floruit
}

func (d *FlatPerson) GetComment() string {
	return d.Comment
}

func (d *FlatPerson) IsDummy() bool {
	return d == nil || d.Dummy
}

func (d FlatPerson) String() string {
	return strtpl.MustEval("/Person {{ if .GetID }}{{ .GetID }}{{ else }}{{ .GetUUID }}{{ end }}: {{ with .GetName }}{{ if not .Empty }}{{ range .First }}{{ . }} {{ end }}{{ .Last }}{{ end }}{{ end }}/", &d)
}

func getOtherPerson(a, b, than Person) Person {
	if a == nil && than == nil {
		return b
	}
	if b == nil && than == nil {
		return a
	}
	if than == nil || a == nil || b == nil {
		return nil
	}
	if b == nil || (a.GetID() == than.GetID()) {
		return b
	}
	if a == nil || (b.GetID() == than.GetID()) {
		return a
	}
	return nil
}

// mergePersonSlices merges two slices of FlatPersons filtering duplicate entries
func mergePersonSlices(sources ...[]Person) []*FlatPerson {
	if len(sources) == 0 {
		return []*FlatPerson{}
	}

	result := []*FlatPerson{}
	for _, source := range sources {
		for _, entry := range source {
			if entry == nil {
				result = append(result, nil)
				continue
			}
			result = append(result, entry.(*FlatPerson))
		}
	}
	return result
}

func deduplicatePersonSlices(source []*FlatPerson) []*FlatPerson {
	var (
		seen    = map[string]struct{}{}
		seenNil = false
		ok      = false
	)
	result := []*FlatPerson{}
	for _, entry := range source {
		if entry != nil {
			if _, ok = seen[entry.GetID()]; ok {
				continue
			}
			seen[entry.GetID()] = struct{}{}
		} else {
			if seenNil {
				continue
			}
			seenNil = true
		}

		result = append(result, entry)
	}
	return result
}
