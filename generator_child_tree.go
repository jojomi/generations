package generations

import (
	"bytes"
	"slices"
	"text/template"
)

func renderFullChildTree(p Person, o RenderTreeOptions) ([]byte, error) {
	return renderChildTree(p, o, NodeTypeG, 0)
}

func renderChildTree(p Person, o RenderTreeOptions, baseNodeType NodeType, level int) ([]byte, error) {
	var (
		err error
	)

	// ignored?
	if isPersonIgnored(p, o) {
		return []byte{}, nil
	}

	data := struct {
		FamilyID string

		G        string
		Parent   string
		Children string
		Unions   string

		SiblingsYounger string
		SiblingsOlder   string
	}{}

	// render children by partner
	partnerList, err := p.GetPartners()
	if err != nil {
		return []byte{}, err
	}
	partners := nonIgnored(partnerList, o)

	// unions (other partners)
	if partners.Count() > 0 {
		var unionBuffer bytes.Buffer
		for i, partner := range partners.GetPersons() {
			// special case: first partner -> no union allowed!
			if i == 0 {
				uData, err := renderUnionData(p, partner, o, level)
				if err != nil {
					return nil, err
				}
				data.Parent = uData.Parent
				data.Children = uData.Children
				continue
			}

			data, err := renderUnion(p, partner, o, level)
			if err != nil {
				return nil, err
			}

			unionBuffer.Write(data)
			unionBuffer.WriteString("\n")
		}
		data.Unions = unionBuffer.String()
	}

	// obey AddPhantomChildForEmptyFamilies option if set
	isLeafLevel := level >= o.totalChildTreeDepth
	addPhantomChildForEmptyFamilies := o.AddPhantomChildForEmptyFamilies != nil && *o.AddPhantomChildForEmptyFamilies
	if addPhantomChildForEmptyFamilies && !isLeafLevel && data.Children == "" && data.Unions == "" && data.Parent == "" {
		data.Children = "c[phantom]{}" // for debugging: ,subtree box={show bounding box}
	}

	// render g node for anchor
	mainNodeType := baseNodeType
	if data.Children != "" || data.Unions != "" || data.Parent != "" {
		mainNodeType = NodeTypeG
	}
	opts := *o.RenderPersonOptions
	opts.NodeType = mainNodeType
	opts = *opts.HideImageByLevel(o, -level)
	if level == 0 && !opts.HideRootNodeHighlighting {
		rootNodeKey := "rootnode"
		p.AddAttribute(rootNodeKey)
	}

	g, err := renderPerson(p, opts)
	if err != nil {
		return nil, err
	}
	data.G = string(g)
	if !o.HideFamilyIDs {
		data.FamilyID = "family-" + p.GetID()
	}

	if data.Children == "" && data.Unions == "" && data.Parent == "" && level > 0 {
		return []byte(data.G), nil
	}

	templateFile := o.TemplateFilenameChildTree
	result, err := RenderTemplateFile(templateFile, data, template.FuncMap{})
	if err != nil {
		return []byte{}, err
	}
	return withoutEmptyLines(result), nil
}

type unionData struct {
	FamilyID string
	Parent   string
	Children string
}

func renderUnionData(person, partner Person, o RenderTreeOptions, level int) (unionData, error) {
	var buffer bytes.Buffer
	children, err := person.GetChildrenWith(partner)
	if err != nil {
		return unionData{}, err
	}

	var data unionData

	if !partner.IsDummy() && level < o.MaxChildPartnersGenerations {
		opts := *o.RenderPersonOptions
		opts.NodeType = NodeTypeP
		opts = *opts.HideImageByLevel(o, -level)
		parentData, err := renderPerson(partner, opts)
		if err != nil {
			return data, err
		}
		data.Parent = string(parentData)
		if !o.HideFamilyIDs {
			data.FamilyID = "family-" + partner.GetID()
		}
	}

	if level < o.MaxChildGenerations {
		for _, child := range children.GetPersons() {
			if child == nil {
				continue
			}
			// recursive call
			childData, err := renderChildTree(child, o, NodeTypeC, level+1)
			if err != nil {
				return data, err
			}
			buffer.Write(childData)
			buffer.WriteString("\n")
		}
		data.Children = buffer.String()
	}

	return data, nil
}

func renderUnion(person, partner Person, o RenderTreeOptions, level int) ([]byte, error) {
	data, err := renderUnionData(person, partner, o, level)
	if err != nil {
		return []byte{}, nil
	}

	templateFile := o.TemplateFilenameUnionTree
	result, err := RenderTemplateFile(templateFile, data, template.FuncMap{})
	if err != nil {
		return []byte{}, err
	}
	return withoutEmptyLines(result), nil
}

func isPersonIgnored(p Person, oTree RenderTreeOptions) bool {
	return slices.ContainsFunc(oTree.IgnoreIDs, p.MatchesSearch)
}

func nonIgnored(personList *PersonList, oTree RenderTreeOptions) *PersonList {
	result := NewPersonList()
	for _, person := range personList.GetPersons() {
		if isPersonIgnored(person, oTree) {
			continue
		}
		result.AddPerson(person)
	}
	return &result
}
