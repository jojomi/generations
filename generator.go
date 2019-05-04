package generations

import (
	"bytes"
	"io/ioutil"
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
	childTree, err := renderFullChildTree(p, o)
	if err != nil {
		return []byte{}, errors.Annotate(err, "could not render child subtree")
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
		siblings := NewPersonList(nil)
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
	})
	if err != nil {
		return []byte{}, errors.Annotatef(err, "could not render genealogytree template %s for person %s", o.TemplateFilenameTree, p)
	}
	return withoutEmptyLines(result), nil
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
	})
	if err != nil {
		return []byte{}, errors.Annotatef(err, "could not render template %s for person %s", o.TemplateFilename, p)
	}
	return withoutPercentageLines(withoutEmptyLines(result)), nil
}

func withoutEmptyLines(input []byte) []byte {
	var re = regexp.MustCompile(`\n(\s*\n)+`)
	return re.ReplaceAll(input, []byte("\n"))
}

func withoutEmptyLinesString(input string) string {
	var re = regexp.MustCompile(`\n(\s*\n)+`)
	return re.ReplaceAllString(input, "\n")
}

func withoutPercentageLines(input []byte) []byte {
	var re = regexp.MustCompile(`\n(\s*%\s*)+\n`)
	return re.ReplaceAll(input, []byte("\n"))
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

func latexify(input int) string {
	if input < 0 {
		return strings.Repeat("A", -input)
	}
	return strings.Repeat("I", input)
}

func RenderTemplateFile(filename string, data interface{}) ([]byte, error) {
	templateContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}
	result, err := strtpl.EvalWithFuncMap(string(templateContent), template.FuncMap{
		"noEmptyLines":           withoutEmptyLines,
		"noEmptyLinesString":     withoutEmptyLinesString,
		"toString":               toString,
		"join":                   strings.Join,
		"getFilteredStringSlice": getFilteredStringSlice,
		"latexify":               latexify,
	}, data)
	if err != nil {
		return []byte{}, err
	}
	return []byte(result), nil
}
