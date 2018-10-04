package generations

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderGenealogytree(t *testing.T) {
	renderOptions := RenderTreeOptions{
		GraphType:     GraphTypeSandclock,
		GenderOrder:   GenderOrderMaleFirst,
		HideFamilyIDs: true,
	}
	addTestTemplates(&renderOptions)
	renderOptions.SetDefaults()
	renderOptions.RenderPersonOptions.HideAllData()
	renderOptions.RenderPersonOptions.HideRootNodeHighlighting = true

	tests := []struct {
		Name                 string
		DatabaseFilename     string
		ID                   string
		MaxParentGenerations int
		MaxChildGenerations  int
		Expected             string
	}{
		// alone
		{
			DatabaseFilename: "single",
			ID:               "gauss",
			Expected: `sandclock{
						child{
							g[id=gauss,]{}
						}
					}`,
		},
		// order of siblings
		{
			DatabaseFilename:     "siblings-order",
			ID:                   "gauss",
			MaxParentGenerations: GenerationsNone,
			Expected: `sandclock{
			c[id=schwester1,]{}
			child{
				g[id=gauss,]{}
			}
			c[id=schwester2,]{}
		}`,
		},
		// grandparents
		{
			DatabaseFilename: "grandparents",
			ID:               "gauss",
			Expected: `sandclock{
	parent{
		g[id=papa,]{}
		p[id=opa,]{}
		p[id=oma,]{}
	}
	p[id=mama,]{}
	child {
		g[id=gauss,]{}
	}
}`,
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
		renderOptions.MaxParentGenerations = test.MaxParentGenerations
		renderOptions.MaxChildGenerations = test.MaxChildGenerations
		result, err := RenderGenealogytree(person, renderOptions)
		assert.Nil(t, err, test.Name)
		assertOutputSemantic(t, test.Expected, string(result), test.Name)
	}
}
