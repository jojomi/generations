package generations

import (
	"os"
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

	index map[string]*FlatPerson
}

func NewMemoryDatabase() *MemoryDatabase {
	db := MemoryDatabase{}
	return &db
}

func (y *MemoryDatabase) ParseYamlFile(filename string) error {
	persons := y.Persons

	data, err := os.ReadFile(filename)
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
		if len(p.Sources) == 0 {
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
	y.index = nil

	// add sources to database
	y.Sources = append(y.Sources, yamlDatabase.Sources...)

	return nil
}

func (y *MemoryDatabase) WriteYamlFile(filename string) error {
	data, err := yaml.Marshal(y)
	if err != nil {
		return errors.Annotate(err, "error marshalling database to yaml")
	}

	err = os.WriteFile(filename, data, 0640)
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
	y.index = nil
	return nil
}

func (y MemoryDatabase) Anonymize() {
	for i, p := range y.Persons {
		yearOfBirth, err := strconv.Atoi(firstLetters(p.Birth.Date, 4))
		if err == nil && yearOfBirth < 1880 {
			continue
		}
		if len(p.Name.First) > 0 {
			if p.Name.Used != "" {
				p.Name = Name{
					First: []string{firstLetters(p.Name.Used, 1) + "."},
				}
			} else {
				p.Name = Name{
					First: []string{firstLetters(p.Name.First[0], 1) + "."},
				}
			}
		} else {
			p.Name = Name{}
		}
		p.Birth.Place = ""
		p.Birth.Date = firstLetters(p.Birth.Date, 4)
		p.Death.Place = ""
		p.Death.Date = firstLetters(p.Death.Date, 4)
		p.Baptism = DatePlace{}
		p.Burial = DatePlace{}
		p.Jobs = make([]Job, 0)
		for j, r := range p.Partners {
			r.Engagement = DatePlace{}
			r.Marriage.Date = firstLetters(r.Marriage.Date, 4)
			r.Divorce.Date = firstLetters(r.Divorce.Date, 4)
			p.Partners[j] = r
		}
		p.Residences = make([]Residence, 0)

		p.BiographyElements = make([]BiographyElement, 0)
		p.Comments = make([]Comment, 0)
		p.Sources = make([]Reference, 0)
		y.Persons[i] = p
	}
}

func (y *MemoryDatabase) ensureIndex() {
	if y.index != nil {
		return
	}
	y.index = make(map[string]*FlatPerson, len(y.Persons)*2)
	for _, p := range y.Persons {
		if id := p.GetID(); id != "" {
			y.index[id] = p
		}
		if uuid := p.GetUUID(); uuid != "" {
			y.index[uuid] = p
		}
	}
}

func (y *MemoryDatabase) Get(search string) (Person, error) {
	y.ensureIndex()
	if p, ok := y.index[search]; ok {
		return p, nil
	}
	for _, p := range y.Persons {
		if p.MatchesSearch(search) {
			return p, nil
		}
	}
	return nil, errors.Errorf("person not found for search '%s'", search)
}

func (y *MemoryDatabase) GetByID(ID string) (Person, error) {
	y.ensureIndex()
	if p, ok := y.index[ID]; ok {
		return p, nil
	}
	return nil, errors.Errorf("person not found for ID %s", ID)
}

func firstLetters(input string, count int) string {
	runes := []rune(input)
	if count > len(runes) {
		count = len(runes)
	}
	return string(runes[:count])
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
		year := fourDigitYearRegexp.FindString(birth.Date)
		id += year
	}
	return id, nil
}

var fourDigitYearRegexp = regexp.MustCompile(`[12]\d{3}`)

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
