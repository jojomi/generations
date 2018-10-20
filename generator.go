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
		mom, dad := p.GetParents()
		siblings := []Person{}
		if !dad.IsDummy() {
			siblings = dad.GetChildrenWith(mom)
		} else if !mom.IsDummy() {
			siblings = mom.GetChildrenWith(NewDummyFlatPerson())
		}

		// apply ignore rules
		siblings = nonIgnored(siblings, o)

		// split older and younger
		younger, older := SplitPersons(siblings, p)
		opts := *o.RenderPersonOptions
		opts.NodeType = NodeTypeC
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

func renderPersonSlice(persons []Person, renderPersonOptions RenderPersonOptions) (string, error) {
	var outputBuffer bytes.Buffer
	for _, person := range persons {
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

func RenderTemplateFile(filename string, data interface{}) ([]byte, error) {
	templateContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}
	result, err := strtpl.EvalWithFuncMap(string(templateContent), template.FuncMap{
		"noEmptyLines":       withoutEmptyLines,
		"noEmptyLinesString": withoutEmptyLinesString,
		"toString":           toString,
		"join":               strings.Join,
	}, data)
	if err != nil {
		return []byte{}, err
	}
	return []byte(result), nil
}
