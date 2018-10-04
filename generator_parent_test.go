package generations

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderFullParentTree(t *testing.T) {
	renderOptions := RenderTreeOptions{
		GraphType:     GraphTypeParent,
		GenderOrder:   GenderOrderMaleFirst,
		HideFamilyIDs: true,
	}
	addTestTemplates(&renderOptions)
	renderOptions.SetDefaults()
	renderOptions.RenderPersonOptions.HideAllData()

	tests := []struct {
		Name                 string
		DatabaseFilename     string
		ID                   string
		Headless             bool
		GenderOrder          GenderOrder
		MaxParentGenerations int
		Expected             string
	}{
		// alone
		{
			DatabaseFilename: "single",
			ID:               "gauss",
			Expected:         `parent{g[id=gauss,]{}}`,
		},
		// with parents
		{
			DatabaseFilename: "parents",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
p[id=papa,]{}
p[id=mama,]{}
}`,
		},
		{
			DatabaseFilename: "single-parent",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
p[id=mama,]{}
}`,
		},
		{
			DatabaseFilename: "parents-extra-data",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
p[id=papa,]{}
p[id=mama,]{}
}`,
		},
		{
			DatabaseFilename: "grandparents",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
parent{
	g[id=papa,]{}
	p[id=opa,]{}
	p[id=oma,]{}
}
p[id=mama,]{}
}`,
		},
		{
			DatabaseFilename: "multi-generation-parents",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
parent{
	g[id=mama,]{}
	parent{
		g[id=oma,]{}
		parent{
			g[id=uroma,]{}
			p[id=ururoma,]{}
		}
	}
}
}`,
		},

		// MaxParentGenerations
		{
			DatabaseFilename:     "multi-generation-parents",
			ID:                   "gauss",
			MaxParentGenerations: 2,
			Expected: `parent{
g[id=gauss,]{}
parent{
	g[id=mama,]{}
	p[id=oma,]{}
}
}`,
		},
		{
			Name:                 "MaxParentGenerations",
			DatabaseFilename:     "multi-generation-parents",
			ID:                   "gauss",
			Headless:             false,
			MaxParentGenerations: GenerationsNone,
			Expected: `parent{
g[id=gauss,]{}
}`,
		},

		// headless tests
		{
			DatabaseFilename: "parents",
			ID:               "gauss",
			Headless:         true,
			Expected: `
p[id=papa,]{}
p[id=mama,]{}
`,
		},

		// siblings
		{
			DatabaseFilename: "parents-siblings",
			ID:               "gauss",
			Expected: `parent{
g[id=gauss,]{}
c[id=schwester,]{}
p[id=papa,]{}
p[id=mama,]{}
}`,
		},

		// gender order
		{
			Name:             "parents-gender-order-male-first",
			DatabaseFilename: "parents",
			ID:               "gauss",
			GenderOrder:      GenderOrderMaleFirst,
			Expected: `parent{
g[id=gauss,]{}
p[id=papa,]{}
p[id=mama,]{}
}`,
		},
		{
			Name:             "parents-gender-order-female-first",
			DatabaseFilename: "parents",
			ID:               "gauss",
			GenderOrder:      GenderOrderFemaleFirst,
			Expected: `parent{
g[id=gauss,]{}
p[id=mama,]{}
p[id=papa,]{}
}`,
		},
	}

	for _, test := range tests {
		database := NewMemoryDatabase()
		err := database.ParseYamlFile(filepath.Join("testdata", "database", test.DatabaseFilename+".yml"))
		assert.Nil(t, err)
		person, err := database.GetByID(test.ID)
		assert.Nil(t, err)
		renderOptions.MaxParentGenerations = test.MaxParentGenerations
		renderOptions.GenderOrder = 0
		if test.GenderOrder != 0 {
			renderOptions.GenderOrder = test.GenderOrder
		}
		renderOptions.SetDefaults()
		result, err := renderFullParentTree(person, renderOptions, test.Headless)
		if err != nil {
			fmt.Println(err)
		}
		assert.Nil(t, err)
		if test.Name == "" {
			test.Name = test.DatabaseFilename
		}
		if !assertOutputSemantic(t, test.Expected, string(result), test.Name) {
			break
		}
	}
}
