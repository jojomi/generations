package generations

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderFullChildTree(t *testing.T) {
	renderOptions := RenderTreeOptions{
		GraphType:   GraphTypeChild,
		GenderOrder: GenderOrderMaleFirst,
	}
	addTestTemplates(&renderOptions)
	renderOptions.SetDefaults()
	renderOptions.RenderPersonOptions.HideAllData()
	renderOptions.RenderPersonOptions.HideRootNodeHighlighting = true

	tests := []struct {
		Name                string
		DatabaseFilename    string
		ID                  string
		MaxChildGenerations int
		Expected            string
	}{
		// alone
		{
			DatabaseFilename: "single",
			ID:               "gauss",
			Expected:         `child{g[id=gauss,]{}}`,
		},
		// children
		{
			DatabaseFilename: "children",
			ID:               "gauss",
			Expected: `child{
				g[id=gauss,]{}
				union {
					p[id=frau-gauss,]{}
					c[id=sohn,]{}
					c[id=tochter,]{}
				}
			}`,
		},
		{
			DatabaseFilename: "children-order",
			ID:               "gauss",
			Expected: `child{
				g[id=gauss,]{}
				union {
					p[id=frau-gauss,]{}
					c[id=tochter1,]{}
					c[id=sohn,]{}
					c[id=tochter2,]{}
				}
			}`,
		},

		// children with partner
		{
			DatabaseFilename: "children-partner",
			ID:               "gauss",
			Expected: `child{
				g[id=gauss,]{}
				union {
					child {
						g[id=tochter,]{}
						union {
							p[id=schwiegersohn,]{}
							c[id=enkelin,]{}
						}
					}
				}
			}`,
		},

		// child without known partner
		{
			DatabaseFilename: "single-parent",
			ID:               "mama",
			Expected: `child{
				g[id=mama,]{}
				union {
					c[id=gauss,]{}
				}
			}`,
		},

		// union family
		{
			DatabaseFilename: "children-union",
			ID:               "gauss",
			Expected: `child{
				g[id=gauss,]{}
				union {
					p[id=frau-gauss,]{}
					c[id=sohn,]{}
					c[id=tochter,]{}
				}
				union {
					p[id=frau-gauss-zwei,]{}
					c[id=sohn-zwei,]{}
				}
			}`,
		},

		// MaxChildGenerations
		{
			Name:                "MaxChildGenerations",
			DatabaseFilename:    "children",
			ID:                  "gauss",
			MaxChildGenerations: GenerationsNone,
			Expected: `child{
				g[id=gauss,]{}
				union {
					p[id=frau-gauss,]{}
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
		assert.Nil(t, err)
		person, err := database.GetByID(test.ID)
		assert.Nil(t, err)
		renderOptions.MaxChildGenerations = test.MaxChildGenerations
		renderOptions.SetDefaults()
		result, err := renderFullChildTree(person, renderOptions)
		if err != nil {
			fmt.Println(err)
		}
		assert.Nil(t, err)
		assertOutputSemantic(t, test.Expected, string(result), fmt.Sprintf("Test %s failed", test.Name))
	}
}
