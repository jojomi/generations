package generations

import (
	"io/ioutil"
	"math"
	"regexp"
	"strconv"

	"github.com/juju/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type Database interface {
	Get(search string) (Person, error)
	GetByID(ID string) (Person, error)
	MakeIDs(f func(p Person, d Database) error) error
}

type YamlDatabase struct {
	Persons        []*FlatPerson `yaml:",omitempty"`
	Sources        []Source      `yaml:",omitempty"`
	DefaultSources []Reference   `yaml:"default-sources,omitempty"`
}

type MemoryDatabase struct {
	Persons []*FlatPerson
	Sources []Source // TODO Familybook sollte Source heißen und Source eher Reference
}

func NewMemoryDatabase() *MemoryDatabase {
	db := MemoryDatabase{}
	return &db
}

func (y *MemoryDatabase) ParseYamlFile(filename string) error {
	persons := y.Persons

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Annotatef(err, "error reading yaml database %s", filename)
	}
	var yamlDatabase YamlDatabase
	err = yaml.UnmarshalStrict(data, &yamlDatabase)
	if err != nil {
		return errors.Annotatef(err, "syntax error reading yaml database %s", filename)
	}

	// augment: auto ID, set DB handle
	checkDuplicatesOnImport := true
	for _, p := range yamlDatabase.Persons {
		p.Database = y

		// add default sources if no main source is given for the person
		if p.Sources == nil || len(p.Sources) == 0 {
			p.Sources = yamlDatabase.DefaultSources
		}

		// check if duplicate
		if checkDuplicatesOnImport {
			if existing, err := y.Get(p.GetBestID()); err == nil {
				log.Info().Str("person", p.String()).Msg("duplicate person")

				// try to make connections between datasets
				if existing.GetRawDad() == "" {
					existing.SetRawDad(p.GetRawDad())
				}

				if existing.GetRawMom() == "" {
					existing.SetRawMom(p.GetRawMom())
				}

				continue
			}
		}
		log.Trace().Str("person", p.String()).Msg("adding person")
		persons = append(persons, p)
	}
	y.Persons = persons

	// add sources to database
	y.Sources = append(y.Sources, yamlDatabase.Sources...)

	return nil
}

func (y *MemoryDatabase) WriteYamlFile(filename string) error {
	data, err := yaml.Marshal(y)
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

func (y MemoryDatabase) Anonymize() {
	for i, p := range y.Persons {
		yearOfBirth, err := strconv.Atoi(first(p.Birth.Date, 4))
		if err == nil && yearOfBirth < 1880 {
			continue
		}
		if len(p.Name.First) > 0 {
			if p.Name.Used != "" {
				p.Name = Name{
					First: []string{first(p.Name.Used, 1) + "."},
				}
			} else {
				p.Name = Name{
					First: []string{first(p.Name.First[0], 1) + "."},
				}
			}
		} else {
			p.Name = Name{}
		}
		p.Birth.Place = ""
		p.Birth.Date = first(p.Birth.Date, 4)
		p.Death.Place = ""
		p.Death.Date = first(p.Death.Date, 4)
		p.Baptism = DatePlace{}
		p.Burial = DatePlace{}
		p.Jobs = make([]Job, 0)
		for j, r := range p.Partners {
			r.Engagement = DatePlace{}
			r.Marriage.Date = first(r.Marriage.Date, 4)
			r.Divorce.Date = first(r.Divorce.Date, 4)
			p.Partners[j] = r
		}
		p.Residences = make([]Residence, 0)

		p.BiographyElements = make([]BiographyElement, 0)
		p.Comments = make([]Comment, 0)
		p.Sources = make([]Reference, 0)
		y.Persons[i] = p
	}
}

func (y MemoryDatabase) Get(search string) (Person, error) {
	for _, p := range y.Persons {
		if p.MatchesSearch(search) {
			return p, nil
		}
	}
	return nil, errors.Errorf("person not found for search '%s'", search)
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
	if len(name.Used) > 0 && name.Used != "" && name.Used != "?" {
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
		year := re.FindString(birth.Date)
		id += year
	}
	return id, nil
}

func GetFourFourBirthNameID(p Person) (string, error) {
	var id string
	name := p.GetName()
	if name.Birth == "" || name.Birth == "?" {
		return "", nil
	}
	if name.Last != "" {
		id += firstLetters(name.Last, 4)
	}
	if len(name.Used) > 0 && name.Used != "" && name.Used != "?" {
		id += firstLetters(name.Used, 4)
	} else {
		if len(name.First) > 0 && name.First[0] != "" {
			id += firstLetters(name.First[0], 4)
		}
	}
	id += firstLetters(name.Birth, 4)
	return id, nil
}

func GetLastFirstNameID(p Person) (string, error) {
	var id string
	name := p.GetName()
	if name.Last != "" {
		id += name.Last
	}
	if len(name.Used) > 0 && name.Used != "" && name.Used != "?" {
		id += name.Used
	} else {
		if len(name.First) > 0 && name.First[0] != "" {
			id += name.First[0]
		}
	}
	if id == "??" {
		return "", nil
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

func first(input string, count int) string {
	if len(input) <= count {
		return input
	}
	return string([]rune(input[0:count]))
}
