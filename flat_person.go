package generations

import (
	"sort"

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

func (d *FlatPerson) MatchesSearch(search string) bool {
	if d.GetID() == search {
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

func (d *FlatPerson) GetBurial() DatePlace {
	return d.Burial
}

func (d *FlatPerson) GetParents() (Person, Person) {
	return d.GetMom(), d.GetDad()
}

func (d *FlatPerson) GetMom() Person {
	mom, err := d.Database.GetByID(d.Mom)
	if err != nil {
		return NewDummyFlatPerson()
	}
	return mom
}

func (d *FlatPerson) GetDad() Person {
	dad, err := d.Database.GetByID(d.Dad)
	if err != nil {
		return NewDummyFlatPerson()
	}
	return dad
}

func (d *FlatPerson) GetPartners() []Person {
	result := []Person{}

	// find explicit partners
	for _, person := range d.Database.Persons {
		for _, partner := range person.Partners {
			if partner.PartnerID == d.GetID() {
				result = append(result, person)
			}
		}
	}

	// find partners through common children
	childrenPartners := d.GetChildrenParents()

	// merge results
	intermediateResult := deduplicatePersonSlices(mergePersonSlices(childrenPartners, result))

	finalResult := make([]Person, len(intermediateResult))
	for i, p := range intermediateResult {
		finalResult[i] = p
	}
	return finalResult
}

func (d *FlatPerson) GetChildrenParents() []Person {
	result := []Person{}
	parentsSeen := make(map[string]struct{}, 0)
	var (
		mom       Person
		dad       Person
		candidate Person
		ok        bool
	)
	for _, child := range d.GetChildren() {
		mom, dad = child.GetParents()

		// if this FlatPerson is mom or dad the other one is a candidate parent to be returned
		if mom != nil && mom.GetID() == d.GetID() {
			candidate = dad
		}
		if dad != nil && dad.GetID() == d.GetID() {
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
	return result
}

func (d *FlatPerson) GetChildren() []Person {
	result := []Person{}

	var (
		mom Person
		dad Person
	)
	for _, child := range d.Database.Persons {
		mom = child.GetMom()
		if mom != nil && mom.GetID() == d.GetID() {
			result = append(result, child)
			continue
		}
		dad = child.GetDad()
		if dad != nil && dad.GetID() == d.GetID() {
			result = append(result, child)
			continue
		}
	}

	return result
}

func (d *FlatPerson) GetChildrenWith(partner Person) []Person {
	result := []Person{}
	for _, child := range d.GetChildren() {
		otherParent := getOtherPerson(child.GetMom(), child.GetDad(), d)
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

	return result
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
	return strtpl.MustEval("/Person {{ .GetID }}: {{ with .GetName }}{{ if not .Empty }}{{ range .First }}{{ . }} {{ end }}{{ .Last }}{{ end }}{{ end }}/", &d)
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
