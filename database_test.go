package generations

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYamlFile(t *testing.T) {
	db := NewMemoryDatabase()
	assert.Len(t, db.Persons, 0)

	// error conditions
	err := db.ParseYamlFile("testdata/database/_invalid_filename_.yml")
	assert.NotNil(t, err)
	assert.Len(t, db.Persons, 0)
	err = db.ParseYamlFile("testdata/database/invalid.yml")
	assert.NotNil(t, err)
	assert.Len(t, db.Persons, 0)

	// successful reading and parsing
	err = db.ParseYamlFile("testdata/database/single.yml")
	assert.Nil(t, err)
	assert.Len(t, db.Persons, 1)
}

func TestGet(t *testing.T) {
	db := NewMemoryDatabase()
	db.ParseYamlFile("testdata/database/single-full-details.yml")

	// search by ID
	p, err := db.Get("gauss")
	assert.Nil(t, err)
	assert.Equal(t, "gauss", p.GetID())

	// search by name (first + last)
	p, err = db.Get("Carl Gauss")
	assert.Nil(t, err)
	assert.Equal(t, "gauss", p.GetID())

	// no match
	p, err = db.Get("Siegfried Gauss")
	assert.NotNil(t, err)
}

func TestWriteYamlFile(t *testing.T) {
	db := NewMemoryDatabase()
	tempFile, _ := ioutil.TempFile("", "generations-test-")
	tempFileName := tempFile.Name()
	defer os.Remove(tempFileName)
	db.WriteYamlFile(tempFileName)
}

func TestMakeIDs(t *testing.T) {
	db := NewMemoryDatabase()
	p1 := &FlatPerson{
		ID: "manual",
	}
	p2 := &FlatPerson{
		Name: Name{
			First: []string{"Heinz"},
			Last:  "Müller",
		},
		Birth: DatePlace{Date: "1982-06-08"},
	}
	p3 := &FlatPerson{
		Name: Name{
			Last: "Noe",
		},
		Birth: DatePlace{Date: "um 1826"},
	}
	p4 := &FlatPerson{
		Name: Name{
			First: []string{"Andreas", "Wolfgang"},
		},
	}
	db.Persons = []*FlatPerson{
		p1, p2, p3, p4,
	}
	err := db.MakeIDs(FourFourYearIDFunc)
	assert.Nil(t, err)
	assert.Equal(t, "manual", p1.ID)
	assert.Equal(t, "MüllHein1982", p2.ID)
	assert.Equal(t, "Noe1826", p3.ID)
	assert.Equal(t, "Andr", p4.ID)
}
