package generations

import (
	"bytes"

	"github.com/juju/errors"
)

func renderFullParentTree(p Person, o RenderTreeOptions, headless bool) ([]byte, error) {
	return renderParentTree(p, o, NodeTypeG, 0, headless)
}

func renderParentTree(p Person, o RenderTreeOptions, baseNodeType NodeType, level int, headless bool) ([]byte, error) {
	// ignored?
	if isPersonIgnored(p, o) {
		return []byte{}, nil
	}
	if p.IsDummy() {
		return []byte{}, nil
	}

	data := struct {
		FamilyID        string
		G               string
		Parents         string
		SiblingsYounger string
		SiblingsOlder   string
	}{}

	// render parents
	if level < o.MaxParentGenerations {
		mom, err := p.GetMom()
		if err!=nil{
			return []byte{}, err
		}
		dad, err := p.GetDad()
		if err!=nil{
			return []byte{}, err
		}
		var parents []Person

		switch o.GenderOrder {
		case GenderOrderFemaleFirst:
			parents = []Person{mom, dad}
		case GenderOrderMaleFirst:
			parents = []Person{dad, mom}
		}
		var buffer bytes.Buffer
		for _, parent := range parents {
			if parent.IsDummy() {
				continue
			}
			// recursive call
			parentData, err := renderParentTree(parent, o, NodeTypeP, level+1, false)
			if err != nil {
				return nil, err
			}
			buffer.Write(parentData)
			buffer.WriteString("\n")
		}
		data.Parents = buffer.String()

		// render siblings
		if level <= o.MaxParentSiblingsGenerations {
			var siblings []Person
			if !mom.IsDummy() {
				siblings, err = mom.GetChildrenWith(dad)
				if err != nil {
					return []byte{}, err
				}
			} else if !dad.IsDummy() {
				siblings, err = dad.GetChildrenWith(NewDummyFlatPerson())
				if err != nil {
					return []byte{}, err
				}
			}

			// apply ignore rules
			siblings = nonIgnored(siblings, o)

			younger, older := SplitPersons(siblings, p)
			opts := *o.RenderPersonOptions
			opts.NodeType = NodeTypeC
			siblingsOlder, err := renderPersonSlice(older, opts)
			if err != nil {
				return []byte{}, err
			}
			siblingsYounger, err := renderPersonSlice(younger, opts)
			if err != nil {
				return []byte{}, err
			}
			data.SiblingsOlder = siblingsOlder
			data.SiblingsYounger = siblingsYounger
		}
	}

	// render g node for anchor
	mainNodeType := baseNodeType
	if data.Parents != "" {
		mainNodeType = NodeTypeG
	}
	opts := *o.RenderPersonOptions
	opts.NodeType = mainNodeType
	g, err := renderPerson(p, opts)
	if err != nil {
		return nil, errors.Annotatef(err, "could not render g node for %s", p)
	}
	data.G = string(g)
	if !o.HideFamilyIDs {
		data.FamilyID = "family-" + p.GetID()
	}

	if data.Parents == "" && level > 0 {
		return []byte(data.G), nil
	}

	var templateFile string
	if headless {
		templateFile = o.TemplateFilenameParentTreeHeadless
	} else {
		templateFile = o.TemplateFilenameParentTree
	}
	result, err := RenderTemplateFile(templateFile, data)
	if err != nil {
		return []byte{}, errors.Annotatef(err, "could not render template based on file %s", templateFile)
	}
	return withoutEmptyLines(result), nil
}
