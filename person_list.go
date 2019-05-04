package generations

import "sort"

// PersonList is a model for collection of persons
type PersonList struct {
	persons  []Person
	database Database
}

func NewPersonList(database Database) PersonList {
	return PersonList{
		persons:  []Person{},
		database: database,
	}
}

func NewPersonListByIDs(database Database, IDs []string) (PersonList, error) {
	p := NewPersonList(database)
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

func (p *PersonList) RemoveDuplicates() *PersonList {
	result := []Person{}
	seenMap := map[string]struct{}{}
	for _, person := range p.GetPersons() {
		key := person.GetID()
		if _, ok := seenMap[key]; ok {
			continue
		}
		result = append(result, person)
		seenMap[key] = struct{}{}
	}
	p.persons = result
	return p // for chainability
}

func (p *PersonList) Invert() *PersonList {
	// TODO implement
	return p // for chainability
}

func (p *PersonList) GetPartners() (PersonList, error) {
	result := NewPersonList(p.database)
	var (
		partners PersonList
		err      error
	)
	for _, person := range p.GetPersons() {
		partners, err = person.GetPartners()
		if err != nil {
			return result, err
		}
		result.AddList(&partners)
	}
	return result, nil
}

func (p *PersonList) GetChildren() (PersonList, error) {
	result := NewPersonList(p.database)
	var (
		children PersonList
		err      error
	)
	for _, person := range p.GetPersons() {
		children, err = person.GetChildren()
		if err != nil {
			return result, err
		}
		result.AddList(&children)
	}
	return result, nil
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

func (p *PersonList) SortPersons(sortFunc func(i, j int) bool) *PersonList {
	sort.Slice(p.persons, sortFunc)
	return p // for chainability
}
