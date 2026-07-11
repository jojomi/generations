package generations

import (
	"sort"
	"strings"
)

// PersonList is a model for collection of persons
type PersonList struct {
	persons  []Person
	database Database
	sortFunc func(a, b Person) bool
}

type AttributePersons struct {
	Attribute string
	Persons   *PersonList
}

func NewPersonList() PersonList {
	return PersonList{
		persons: []Person{},
	}
}

func NewPersonListWithDatabase(database Database) PersonList {
	return PersonList{
		persons:  []Person{},
		database: database,
	}
}

func NewPersonListByIDs(database Database, IDs []string) (PersonList, error) {
	p := NewPersonListWithDatabase(database)
	for _, id := range IDs {
		person, err := database.GetByID(id)
		if err != nil {
			return p, err
		}
		p.persons = append(p.persons, person)
	}
	return p, nil
}

func (p *PersonList) GetPersons() []Person {
	return p.persons
}

func (p *PersonList) Count() int {
	return len(p.persons)
}

// GetTreePositionID calculates and returns the TreePositionID. The ID is created by concatenating all IDs in the PersonList with "-".
func (p *PersonList) GetTreePositionID() string {
	var idConcat strings.Builder
	for _, person := range p.GetPersons() {
		idConcat.WriteString("-")
		idConcat.WriteString(person.GetBestID())
	}
	return idConcat.String()
}

func (p *PersonList) GetTreePositionIDWith(person Person) string {
	return p.GetTreePositionID() + "-" + person.GetBestID()
}

func (p *PersonList) Contains(person Person) bool {
	for _, p1 := range p.GetPersons() {
		if p1.Equals(person) {
			return true
		}
	}
	return false
}

func (p *PersonList) HasSameElements(other *PersonList) bool {
	if len(p.GetPersons()) != len(other.GetPersons()) {
		return false
	}
outer:
	for _, p1 := range p.GetPersons() {
		for _, p2 := range other.GetPersons() {
			if p1.GetBestID() == p2.GetBestID() {
				continue outer
			}
		}
		return false
	}
	return true
}

func (p *PersonList) MergeTreePositions(other *PersonList) *PersonList {
outer:
	for _, p1 := range p.GetPersons() {
		for _, p2 := range other.GetPersons() {
			if p1.GetBestID() == p2.GetBestID() {
				for _, id := range p2.GetTreePositionIDs() {
					p1.AddTreePositionID(id)
				}
				continue outer
			}
		}
	}
	return p
}

func (p *PersonList) GetWithMergedTreePositionIDs() *PersonList {
	result := NewPersonList()
	kept := make(map[string]Person, len(p.persons))
	for _, p1 := range p.GetPersons() {
		bestID := p1.GetBestID()
		if existing, ok := kept[bestID]; ok {
			for _, id := range p1.GetTreePositionIDs() {
				existing.AddTreePositionID(id)
			}
			continue
		}
		kept[bestID] = p1
		result.AddPerson(p1)
	}

	return &result
}

func (p *PersonList) GetAttributes() []AttributePersons {
	attrMap := make(map[string]*PersonList, 0)
	attrList := make([]string, 0)

	for _, person := range p.GetPersons() {
		for _, attr := range person.GetAttributes() {
			if _, ok := attrMap[attr]; !ok {
				pl := NewPersonList()
				attrMap[attr] = &pl
				attrList = append(attrList, attr)
			}
			attrMap[attr].AddPerson(person)
		}
	}

	// sort
	sort.Strings(attrList)

	result := make([]AttributePersons, len(attrList))
	for i, attr := range attrList {
		result[i] = AttributePersons{
			Attribute: attr,
			Persons:   attrMap[attrList[i]],
		}
	}

	return result
}

func (p *PersonList) RemoveDuplicatesByBestID() *PersonList {
	result := make([]Person, 0, len(p.persons))
	seen := make(map[string]struct{}, len(p.persons))
	for _, currP := range p.persons {
		bestID := currP.GetBestID()
		if _, ok := seen[bestID]; ok {
			continue
		}
		seen[bestID] = struct{}{}
		result = append(result, currP)
	}
	p.persons = result
	return p
}

func (p *PersonList) Exclude(persons ...Person) {
	result := make([]Person, 0, len(p.GetPersons()))
outer:
	for _, person := range p.GetPersons() {
		for _, fPerson := range persons {
			if person.Equals(fPerson) {
				continue outer
			}
		}
		result = append(result, person)
	}
	p.persons = result
}

func (p *PersonList) ExcludeList(persons *PersonList) {
	p.Exclude(persons.GetPersons()...)
}

func (p *PersonList) Remove(person Person) *PersonList {
	if !p.Contains(person) {
		return p
	}

	persons := p.GetPersons()
	filtered := make([]Person, 0, len(persons))
	for _, pers := range persons {
		if person.Equals(pers) {
			continue
		}
		filtered = append(filtered, pers)
	}

	p.persons = filtered
	return p
}

func (p *PersonList) Invert() *PersonList {
	// TODO implement
	return p // for chainability
}

func (p *PersonList) GetPartners() (PersonList, error) {
	result := NewPersonListWithDatabase(p.database)
	var (
		partners *PersonList
		err      error
	)
	for _, person := range p.GetPersons() {
		partners, err = person.GetPartners()
		if err != nil {
			return result, err
		}
		result.AddList(partners)
	}
	return result, nil
}

func (p *PersonList) GetChildren() (PersonList, error) {
	result := NewPersonListWithDatabase(p.database)
	var (
		children *PersonList
		err      error
	)
	for _, person := range p.GetPersons() {
		children, err = person.GetChildren()
		if err != nil {
			return result, err
		}
		result.AddList(children)
	}
	result.RemoveDuplicatesByBestID()
	return result, nil
}

func (p *PersonList) Diff(otherList *PersonList) (*PersonList, error) {
	result := NewPersonListWithDatabase(p.database)

	for _, person := range p.GetPersons() {
		if otherList.Contains(person) {
			continue
		}
		result.AddPerson(person)
	}

	return &result, nil
}

func (p *PersonList) AddChildren() (*PersonList, error) {
	children, err := p.GetChildren()
	if err != nil {
		return p, err
	}
	p.AddList(&children)
	return p, nil
}

func (p *PersonList) GetParents() (PersonList, error) {
	result := NewPersonListWithDatabase(p.database)
	for _, person := range p.GetPersons() {
		mom, err := person.GetMom()
		if err != nil {
			return result, err
		}
		if !mom.IsDummy() {
			result.AddPerson(mom)
		}
		dad, err := person.GetDad()
		if err != nil {
			return result, err
		}
		if !dad.IsDummy() {
			result.AddPerson(dad)
		}
	}
	result.RemoveDuplicatesByBestID()
	return result, nil
}

func (p *PersonList) AddParents() (*PersonList, error) {
	parents, err := p.GetParents()
	if err != nil {
		return p, err
	}
	p.AddList(&parents)
	return p, nil
}

func (p *PersonList) AddPerson(person Person) *PersonList {
	p.persons = append(p.persons, person)
	return p // for chainability
}

func (p *PersonList) AddList(personList *PersonList) *PersonList {
	for _, person := range personList.GetPersons() {
		p.AddPerson(person)
	}
	return p // for chainability
}

func (p *PersonList) AddSlice(persons []Person) *PersonList {
	for _, person := range persons {
		p.AddPerson(person)
	}
	return p // for chainability
}

func (p *PersonList) GetFiltered(validationFunc func(p Person) bool) *PersonList {
	result := PersonList{}
	for _, person := range p.GetPersons() {
		if !validationFunc(person) {
			continue
		}
		result.AddPerson(person)
	}
	return &result
}

func (p *PersonList) Len() int {
	return len(p.persons)
}

func (p *PersonList) Less(i, j int) bool {
	return p.sortFunc(p.persons[i], p.persons[j])
}

func (p *PersonList) Swap(i, j int) {
	p.persons[i], p.persons[j] = p.persons[j], p.persons[i]
}

func (p *PersonList) SortPersons(sortFunc func(a, b Person) bool) *PersonList {
	p.sortFunc = sortFunc
	sort.Stable(p)
	return p // for chainability
}
