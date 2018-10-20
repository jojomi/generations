package generations

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChildren(t *testing.T) {
	tests := []struct {
		Name             string
		DatabaseFilename string
		ID               string
		ExpectedIDs      []string
	}{
		{
			Name:             "children of dad",
			DatabaseFilename: "children",
			ID:               "gauss",
			ExpectedIDs:      []string{"sohn", "tochter"},
		},
		{
			Name:             "children of mom",
			DatabaseFilename: "children",
			ID:               "frau-gauss",
			ExpectedIDs:      []string{"sohn", "tochter"},
		},
		{
			DatabaseFilename: "single",
			ID:               "gauss",
			ExpectedIDs:      []string{},
		},
	}

	for _, test := range tests {
		if test.Name == "" {
			test.Name = test.DatabaseFilename
		}

		database := NewMemoryDatabase()
		err := database.ParseYamlFile(filepath.Join("testdata", "database", test.DatabaseFilename+".yml"))
		assert.Nil(t, err, test.Name)
		person, err := database.GetByID(test.ID)
		assert.Nil(t, err, test.Name)
		children, err := person.GetChildren()
		assert.Nil(t, err, test.Name)
		assert.Len(t, children, len(test.ExpectedIDs), test.Name)
		childrenIDs := getPersonSliceIDs(children)
		assert.Equal(t, test.ExpectedIDs, childrenIDs, test.Name)
	}
}

func TestGetChildrenParents(t *testing.T) {
	tests := []struct {
		Name             string
		DatabaseFilename string
		ID               string
		ExpectedIDs      []string
	}{
		{
			Name:             "two parents",
			DatabaseFilename: "parents",
			ID:               "mama",
			ExpectedIDs:      []string{"papa"},
		},
		{
			Name:             "two parents inverse",
			DatabaseFilename: "parents",
			ID:               "papa",
			ExpectedIDs:      []string{"mama"},
		},
		{
			Name:             "single parent",
			DatabaseFilename: "single-parent",
			ID:               "mama",
			ExpectedIDs:      []string{"<dummy>"},
		},
		{
			Name:             "no parent",
			DatabaseFilename: "single",
			ID:               "gauss",
			ExpectedIDs:      []string{},
		},
		{
			Name:             "two wifes",
			DatabaseFilename: "two-wifes",
			ID:               "gauss",
			ExpectedIDs:      []string{"frau1", "frau2"},
		},
		{
			Name:             "multiple partners (without marriage)",
			DatabaseFilename: "children-parents",
			ID:               "gauss",
			ExpectedIDs:      []string{"frau1", "frau2", "<dummy>"},
		},
	}

	for _, test := range tests {
		database := NewMemoryDatabase()
		err := database.ParseYamlFile(filepath.Join("testdata", "database", test.DatabaseFilename+".yml"))
		assert.Nil(t, err, test.Name)
		person, err := database.GetByID(test.ID)
		assert.Nil(t, err, test.Name)
		parents, err := person.GetChildrenParents()
		assert.Nil(t, err, test.Name)
		assert.Len(t, parents, len(test.ExpectedIDs), test.Name)
		parentIDs := getPersonSliceIDs(parents)
		assert.Equal(t, test.ExpectedIDs, parentIDs, test.Name)
	}
}

func TestGetPartners(t *testing.T) {
	tests := []struct {
		Name             string
		DatabaseFilename string
		ID               string
		ExpectedIDs      []string
	}{
		{
			DatabaseFilename: "partners",
			ID:               "gauss",
			ExpectedIDs:      []string{"frau-gauss"},
		},
		{
			DatabaseFilename: "partners",
			ID:               "frau-gauss",
			ExpectedIDs:      []string{}, // because there is no partners entry
		},
		{
			DatabaseFilename: "children",
			ID:               "gauss",
			ExpectedIDs:      []string{"frau-gauss"},
		},
		{
			DatabaseFilename: "children",
			ID:               "frau-gauss",
			ExpectedIDs:      []string{"gauss"}, // because there is common children
		},
	}

	for i, test := range tests {
		if test.Name == "" {
			test.Name = "Test #" + strconv.Itoa(i+1) + " (1-based)"
		}
		database := NewMemoryDatabase()
		err := database.ParseYamlFile(filepath.Join("testdata", "database", test.DatabaseFilename+".yml"))
		assert.Nil(t, err, test.Name+": yaml parsing")
		person, err := database.GetByID(test.ID)
		assert.Nil(t, err, test.Name+": database select by ID")
		partners, err := person.GetPartners()
		assert.Nil(t, err, test.Name+": GetPartners no error")
		partnerIDs := getPersonSliceIDs(partners)
		assert.Equal(t, test.ExpectedIDs, partnerIDs, test.Name+": partner IDs")
	}
}
