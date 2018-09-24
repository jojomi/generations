package generations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderPerson(t *testing.T) {
	defaultTemplate := "cmd/templates/person.tpl"
	defaultRenderOptions := RenderPersonOptions{
		TemplateFilename: defaultTemplate,
		NodeType:         NodeTypeG,
	}
	allHiddenRenderPersonOptions := RenderPersonOptions{
		TemplateFilename: defaultTemplate,
		NodeType:         NodeTypeG,
	}
	allHiddenRenderPersonOptions.HideAllData()
	hiddenPlacesRenderPersonOptions := RenderPersonOptions{
		TemplateFilename: defaultTemplate,
		NodeType:         NodeTypeG,
		HidePlaces:       true,
	}
	hiddenMiddleNamesRenderPersonOptions := RenderPersonOptions{
		TemplateFilename: defaultTemplate,
		NodeType:         NodeTypeG,
		HideMiddleNames:  true,
	}

	fullPerson := FlatPerson{
		Attributes:    []string{"dead", "mathematician"},
		Gender:        "male",
		Name:          Name{First: []string{"Johann", "Carl", "Friedrich"}, Last: "Gauss", Birth: "Hauser"},
		Birth:         DatePlace{Date: "1827", Place: "Hannover"},
		Baptism:       DatePlace{Date: "10.10.1827", Place: "Hannover"},
		Death:         DatePlace{Date: "um 1900"},
		Burial:        DatePlace{Place: "Hannover Hauptfriedhof"},
		Floruit:       "Hannover, Berlin",
		Jobs:          "Mathematician, Priest",
		ImageFilename: "images/gauss.jpg",
		Comment:       "Famous.",
		Partners: []FlatRelationship{
			FlatRelationship{
				Engagement: DatePlace{
					Date:  "10/1854",
					Place: "Prag",
				},
				Marriage: DatePlace{
					Date:  "1855",
					Place: "München",
				},
				Divorce: DatePlace{
					Date:  "um 1866",
					Place: "ebd.",
				},
			},
		},
	}

	tests := []struct {
		Name          string
		RenderOptions RenderPersonOptions
		Person        Person
		Expected      string
	}{
		// empty
		{
			RenderOptions: defaultRenderOptions,
			Person:        &FlatPerson{},
			Expected:      "g[]{}",
		},
		// NodeTypes
		{
			RenderOptions: RenderPersonOptions{
				TemplateFilename: defaultTemplate,
				NodeType:         NodeTypeC,
			},
			Person:   &FlatPerson{},
			Expected: "c[]{}",
		},
		{
			RenderOptions: RenderPersonOptions{
				TemplateFilename: defaultTemplate,
				NodeType:         NodeTypeP,
			},
			Person:   &FlatPerson{},
			Expected: "p[]{}",
		},
		// gender
		{
			RenderOptions: defaultRenderOptions,
			Person:        &FlatPerson{Gender: "male"},
			Expected:      "g[]{sex=male,}",
		},
		{
			RenderOptions: defaultRenderOptions,
			Person:        &FlatPerson{Gender: "female"},
			Expected:      "g[]{sex=female,}",
		},
		// full details
		{
			RenderOptions: defaultRenderOptions,
			Person:        &fullPerson,
			Expected: `g[dead,mathematician]{
				sex = male,
				name = { \pref{Johann} \middlename{Carl} \middlename{Friedrich} \surn{Gauss} \surnbirth{Hauser}},
				birth = {1827}{Hannover},
				baptism = {10.10.1827}{Hannover},
				death- = {um 1900},
				burial = {}{Hannover Hauptfriedhof},
				engagement = {10/1854}{Prag},
				marriage = {1855}{München},
				divorce = {um 1866}{ebd.},
				floruit- = {Hannover, Berlin},
				profession = {Mathematician, Priest},
				image = {images/gauss.jpg},
				comment = {Famous.},
			}`,
		},
		// filter
		{
			Name:          "Filter",
			RenderOptions: allHiddenRenderPersonOptions,
			Person:        &fullPerson,
			Expected:      `g[]{}`,
		},
		{
			Name:          "Filter HidePlaces",
			RenderOptions: hiddenPlacesRenderPersonOptions,
			Person: &FlatPerson{
				Birth:   DatePlace{Date: "1821", Place: "München"},
				Baptism: DatePlace{Place: "Hamburg"},
				Death:   DatePlace{Date: "um 1842", Place: "Gera"},
				Partners: []FlatRelationship{
					FlatRelationship{
						Engagement: DatePlace{Date: "1839", Place: "Johannesburg"},
						Marriage:   DatePlace{Place: "Ingolstadt"},
						Divorce:    DatePlace{Place: "Rostock"},
					},
				},
			},
			Expected: `g[]{
				birth- = {1821},
				baptism- = {},
				death- = {um 1842},
				engagement- = {1839},
				marriage- = {},
				divorce- = {},
			}`,
		},
		{
			Name:          "Filter HideMiddleNames",
			RenderOptions: hiddenMiddleNamesRenderPersonOptions,
			Person: &FlatPerson{
				Name: Name{First: []string{"Johann", "Carl", "Friedrich"}, Last: "Gauss", Birth: "Hauser"},
			},
			Expected: `g[]{
				name = {
					\pref{Johann} \surn{Gauss} \surnbirth{Hauser}
				},
			}`,
		},
	}

	for i, test := range tests {
		if test.Name == "" {
			test.Name = fmt.Sprintf("Test %d", i+1)
		}
		result, err := renderPerson(test.Person, test.RenderOptions)
		if err != nil {
			fmt.Println(err)
		}
		assert.Nil(t, err)
		assertOutputSemantic(t, test.Expected, string(result), test.Name)
	}
}
