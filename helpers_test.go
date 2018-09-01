package generations

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertOutputSemantic(t *testing.T, expected, actual, message string) bool {
	compareExpected := semantic(expected)
	compareActual := semantic(actual)
	if compareExpected != compareActual {
		return assert.Equal(t, expected, actual, message)
	}
	return true
}

var collapseWhitespaceRegexp = regexp.MustCompile(`\s+`)

func collapseWhitespace(input string) string {
	return collapseWhitespaceRegexp.ReplaceAllString(input, "")
}

func semantic(input string) string {
	result := collapseWhitespace(input)
	r := strings.NewReplacer(`\xspace`, "", "%", "")
	result = r.Replace(result)
	return result
}

func addTestTemplates(o *RenderTreeOptions) {
	o.TemplateFilenameTree = "cmd/templates/tree.tpl"
	o.TemplateFilenamePerson = "cmd/templates/person.tpl"
	o.TemplateFilenameParentTree = "cmd/templates/parent_tree.tpl"
	o.TemplateFilenameParentTreeHeadless = "cmd/templates/parent_tree_headless.tpl"
	o.TemplateFilenameChildTree = "cmd/templates/child_tree.tpl"
	o.TemplateFilenameUnionTree = "cmd/templates/union_tree.tpl"
}

func getPersonSliceIDs(persons []Person) []string {
	result := make([]string, len(persons))
	for i, person := range persons {
		if person == nil {
			result[i] = "<nil>"
			continue
		}
		result[i] = person.GetID()
	}
	return result
}
