package generations

import (
	"bytes"
)

func renderFullChildTree(p Person, o RenderTreeOptions) ([]byte, error) {
	return renderChildTree(p, o, NodeTypeG, 0)
}

func renderChildrenWithPartner(person, partner Person, o RenderTreeOptions, level int) (string, error) {
	var buffer bytes.Buffer
	children := person.GetChildrenWith(partner)
	for _, child := range children {
		if child == nil {
			continue
		}
		// recursive call
		childData, err := renderChildTree(child, o, NodeTypeC, level+1)
		if err != nil {
			return "", err
		}
		buffer.Write(childData)
		buffer.WriteString("\n")
	}
	return buffer.String(), nil
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
		G        string
		Parent   string
		Children string
		Unions   string

		SiblingsYounger string
		SiblingsOlder   string
	}{}

	// render children by partner
	partners := nonIgnored(p.GetPartners(), o)

	// unions (other partners)
	if len(partners) > 0 {
		var unionBuffer bytes.Buffer
		for _, partner := range partners {
			data, err := renderUnion(p, partner, o, level)
			if err != nil {
				return nil, err
			}
			unionBuffer.Write(data)
			unionBuffer.WriteString("\n")
		}
		data.Unions = unionBuffer.String()
	}

	// render g node for anchor
	mainNodeType := baseNodeType
	if data.Children != "" || data.Unions != "" || data.Parent != "" {
		mainNodeType = NodeTypeG
	}
	opts := *o.RenderPersonOptions
	opts.NodeType = mainNodeType
	if level == 0 && !opts.HideRootNodeHighlighting {
		rootNodeKey := "rootnode"
		p.AddAttribute(rootNodeKey)
	}

	g, err := renderPerson(p, opts)
	if err != nil {
		return nil, err
	}
	data.G = string(g)

	if data.Children == "" && data.Unions == "" && data.Parent == "" && level > 0 {
		return []byte(data.G), nil
	}

	templateFile := o.TemplateFilenameChildTree
	result, err := RenderTemplateFile(templateFile, data)
	if err != nil {
		return []byte{}, err
	}
	return withoutEmptyLines(result), nil
}

func renderUnion(person, partner Person, o RenderTreeOptions, level int) ([]byte, error) {
	var buffer bytes.Buffer
	children := person.GetChildrenWith(partner)

	data := struct {
		Parent   string
		Children string
	}{}

	if !partner.IsDummy() && level < o.MaxChildPartnersGenerations {
		opts := *o.RenderPersonOptions
		opts.NodeType = NodeTypeP
		parentData, err := renderPerson(partner, opts)
		if err != nil {
			return nil, err
		}
		data.Parent = string(parentData)
	}

	if level < o.MaxChildGenerations {
		for _, child := range children {
			if child == nil {
				continue
			}
			// recursive call
			childData, err := renderChildTree(child, o, NodeTypeC, level+1)
			if err != nil {
				return []byte{}, nil
			}
			buffer.Write(childData)
			buffer.WriteString("\n")
		}
		data.Children = buffer.String()
	}

	templateFile := o.TemplateFilenameUnionTree
	result, err := RenderTemplateFile(templateFile, data)
	if err != nil {
		return []byte{}, err
	}
	return withoutEmptyLines(result), nil
}

func isPersonIgnored(p Person, oTree RenderTreeOptions) bool {
	for _, id := range oTree.IgnoreIDs {
		if p.MatchesSearch(id) {
			return true
		}
	}
	return false
}

func nonIgnored(persons []Person, oTree RenderTreeOptions) []Person {
	result := make([]Person, 0, len(persons))
	for _, person := range persons {
		if isPersonIgnored(person, oTree) {
			continue
		}
		result = append(result, person)
	}
	return result
}
