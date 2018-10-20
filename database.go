package generations

import (
	"io/ioutil"
	"math"
	"regexp"

	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

type Database interface {
	Get(search string) (Person, error)
	GetByID(ID string) (Person, error)
	MakeIDs(f func(p Person, d Database) error) error
}

type MemoryDatabase struct {
	Persons []*FlatPerson
}

func NewMemoryDatabase() *MemoryDatabase {
	db := MemoryDatabase{}
	return &db
}

func (y *MemoryDatabase) ParseYamlFile(filename string) error {
	persons := y.Persons
	if persons == nil {
		persons = []*FlatPerson{}
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Annotatef(err, "error reading yaml database %s", filename)
	}
	var yamlPersons []*FlatPerson
	err = yaml.UnmarshalStrict(data, &yamlPersons)
	if err != nil {
		return errors.Annotatef(err, "syntax error reading yaml database %s", filename)
	}

	// TODO augment
	// auto ID
	// set DB handle
	for _, p := range yamlPersons {
		p.Database = y
		persons = append(persons, p)
	}
	y.Persons = persons

	return nil
}

func (y *MemoryDatabase) WriteYamlFile(filename string) error {
	data, err := yaml.Marshal(y.Persons)
	if err != nil {
		return errors.Annotate(err, "error marshalling database to yaml")
	}

	err = ioutil.WriteFile(filename, data, 0640)
	if err != nil {
		return errors.Annotatef(err, "error writing database to %s", filename)
	}

	return nil
}

func (y *MemoryDatabase) MakeIDs(f func(p Person, d Database) error) error {
	var err error
	for i, p := range y.Persons {
		err = f(p, y)
		if err != nil {
			return err
		}
		y.Persons[i] = p
	}
	return nil
}

func (y MemoryDatabase) Get(search string) (Person, error) {
	for _, p := range y.Persons {
		if p.MatchesSearch(search) {
			return p, nil
		}
	}
	return nil, errors.Errorf("person not found for search %s", search)
}

func (y MemoryDatabase) GetByID(ID string) (Person, error) {
	for _, p := range y.Persons {
		if p.MatchesIDUUID(ID) {
			return p, nil
		}
	}
	return nil, errors.Errorf("person not found for ID %s", ID)
}

func firstLetters(input string, count int) string {
	runes := []rune(input)
	return string(runes[:int(math.Min(4.0, float64(len(runes))))])
}

func GetFourFourYearID(p Person) (string, error) {
	var id string
	name := p.GetName()
	if name.Last != "" {
		id += firstLetters(name.Last, 4)
	}
	if len(name.Used) > 0 && name.Used != "" {
		id += firstLetters(name.Used, 4)
	} else {
		if len(name.First) > 0 && name.First[0] != "" {
			id += firstLetters(name.First[0], 4)
		}
	}
	birth := p.GetBirth()
	if !birth.Empty() {
		re, err := regexp.Compile(`[12]\d{3}`) // 4-digit year
		if err != nil {
			return "", err
		}
		id += re.FindString(birth.Date)
	}
	return id, nil
}

func FourFourYearIDFunc(p Person, d Database) error {
	// if the ID is already set, don't you dare touch it!
	if p.GetID() != "" {
		return nil
	}

	id, err := GetFourFourYearID(p)
	if err != nil {
		return err
	}
	p.SetID(id)
	return nil
}
