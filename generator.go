package generations

import (
	"bytes"
	"maps"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/jojomi/strtpl"
	"github.com/juju/errors"
)

func RenderGenealogytree(p Person, o RenderTreeOptions) ([]byte, error) {
	o.SetDefaults()

	var err error
	parentTree, err := renderFullParentTree(p, o, true)
	if err != nil {
		return []byte{}, errors.Annotate(err, "could not render parent subtree")
	}
	o.totalChildTreeDepth = getChildTreeDepth(p, o.MaxChildGenerations)
	childTree, err := renderFullChildTree(p, o)
	if err != nil {
		return []byte{}, errors.Annotate(err, "could not render child subtree for "+p.String())
	}

	var (
		siblingsOlder   string
		siblingsYounger string
	)
	if o.MaxParentSiblingsGenerations != GenerationsNone {
		mom, err := p.GetMom()
		if err != nil {
			return []byte{}, err
		}
		dad, err := p.GetDad()
		if err != nil {
			return []byte{}, err
		}
		s := NewPersonList()
		siblings := &s
		if !dad.IsDummy() {
			siblings, err = dad.GetChildrenWith(mom)
			if err != nil {
				return []byte{}, err
			}
		} else if !mom.IsDummy() {
			siblings, err = mom.GetChildrenWith(NewDummyFlatPerson())
			if err != nil {
				return []byte{}, err
			}
		}

		// apply ignore rules
		siblings = nonIgnored(siblings, o)

		// split older and younger
		younger, older := SplitPersons(siblings, p)
		opts := *o.RenderPersonOptions
		opts.NodeType = NodeTypeC
		opts = *opts.HideImageByLevel(o, 0)
		siblingsOlder, err = renderPersonSlice(older, opts)
		if err != nil {
			return []byte{}, err
		}
		siblingsYounger, err = renderPersonSlice(younger, opts)
		if err != nil {
			return []byte{}, err
		}
	}

	result, err := RenderTemplateFile(o.TemplateFilenameTree, struct {
		ParentTree      string
		ChildTree       string
		SiblingsYounger string
		SiblingsOlder   string
		Options         RenderTreeOptions
	}{
		ParentTree:      string(parentTree),
		ChildTree:       string(childTree),
		SiblingsOlder:   siblingsOlder,
		SiblingsYounger: siblingsYounger,
		Options:         o,
	}, template.FuncMap{})
	if err != nil {
		return []byte{}, errors.Annotatef(err, "could not render genealogytree template %s for person %s", o.TemplateFilenameTree, p)
	}
	return withoutEmptyLines(result), nil
}

func getChildTreeDepth(person Person, maxLevel int) int {
	levels := getChildTreeDepthRecursive(person, 0, 0, maxLevel)
	return levels
}

func getChildTreeDepthRecursive(p Person, level, currMax, filterLevel int) int {
	if level > filterLevel {
		return currMax
	}
	children, err := p.GetChildren()
	if err != nil {
		panic(err)
	}
	if level > currMax {
		currMax = level
	}
	for _, c := range children.GetPersons() {
		childDepth := getChildTreeDepthRecursive(c, level+1, currMax, filterLevel)
		if childDepth > currMax {
			currMax = childDepth
		}
	}
	return currMax
}

func renderPersonSlice(personList PersonList, renderPersonOptions RenderPersonOptions) (string, error) {
	var outputBuffer bytes.Buffer
	for _, person := range personList.GetPersons() {
		personData, err := renderPerson(person, renderPersonOptions)
		if err != nil {
			return "", errors.Annotate(err, "could not render siblings")
		}
		outputBuffer.Write(personData)
	}
	return outputBuffer.String(), nil
}

func renderPerson(p Person, o RenderPersonOptions) ([]byte, error) {
	o = *o.SetDefaults()
	result, err := RenderTemplateFile(o.TemplateFilename, struct {
		Person  Person
		Options RenderPersonOptions
	}{
		Person:  p,
		Options: o,
	}, template.FuncMap{})
	if err != nil {
		return []byte{}, errors.Annotatef(err, "could not render template %s for person %s", o.TemplateFilename, p)
	}
	return withoutPercentageLines(withoutEmptyLines(result)), nil
}

var (
	emptyLinesRegexp      = regexp.MustCompile(`\n(\s*\n)+`)
	percentageLinesRegexp = regexp.MustCompile(`\n(\s*%\s*)+\n`)
)

func withoutEmptyLines(input []byte) []byte {
	return emptyLinesRegexp.ReplaceAll(input, []byte("\n"))
}

func withoutEmptyLinesString(input string) string {
	return emptyLinesRegexp.ReplaceAllString(input, "\n")
}

func withoutPercentageLines(input []byte) []byte {
	return percentageLinesRegexp.ReplaceAll(input, []byte("\n"))
}

func dict(pairs ...interface{}) (map[string]interface{}, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.Errorf("dict requires an even number of arguments")
	}
	d := make(map[string]interface{}, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, errors.Errorf("dict keys must be strings")
		}
		d[key] = pairs[i+1]
	}
	return d, nil
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func getFilteredStringSlice(source, filterElements []string) []string {
	result := make([]string, 0, len(source))
outer:
	for _, elem := range source {
		for _, filterElem := range filterElements {
			if elem == filterElem {
				continue outer
			}
		}
		result = append(result, elem)
	}
	return result
}

func TemplateLatexifyFunc(input string) string {
	r := strings.NewReplacer(
		`"`, ``,
		`_`, `\_`,
		`&`, `\&`,
		`$`, `\$`,
	)
	return r.Replace(input)
}

func RenderTemplateFile(filename string, data interface{}, customFuncMap template.FuncMap) ([]byte, error) {
	templateContent, err := os.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}
	funcMap := template.FuncMap{
		"noEmptyLines":           withoutEmptyLines,
		"noEmptyLinesString":     withoutEmptyLinesString,
		"toString":               toString,
		"join":                   strings.Join,
		"getFilteredStringSlice": getFilteredStringSlice,
		"latexify":               TemplateLatexifyFunc,
		"dict":                   dict,
	}
	maps.Copy(funcMap, customFuncMap)
	result, err := strtpl.EvalWithFuncMap(string(templateContent), funcMap, data)
	if err != nil {
		return []byte{}, err
	}
	return []byte(result), nil
}
