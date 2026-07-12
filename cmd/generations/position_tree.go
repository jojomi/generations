package main

import (
	"fmt"
	"strings"

	"github.com/jojomi/generations"
)

// positionParentTree builds a bare genealogytree parent tree (dots only, no
// person data) from the proband up to targetLevel generations. Only ancestors
// present in the database are included, so branches may end below targetLevel.
// Nodes at targetLevel get the "positionmark" box style (defined in the
// document template). The root node carries a sex field if all root persons
// share a known gender, otherwise it stays neutral.
func positionParentTree(persons []generations.Person, targetLevel int, maternalOrder bool) (string, error) {
	if len(persons) == 0 || targetLevel < 1 {
		return "", nil
	}
	proband := persons[0]

	body, err := positionParentSubtree(proband, 1, targetLevel, maternalOrder)
	if err != nil {
		return "", err
	}
	if body == "" {
		// no known parents, nothing to draw
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("parent{\n")
	sb.WriteString(positionNode("g", commonGender(persons), 0, targetLevel))
	sb.WriteString(body)
	sb.WriteString("}\n")
	return sb.String(), nil
}

// positionParentSubtree returns the parent entries (p or nested parent) for
// the parents of p at the given level.
func positionParentSubtree(p generations.Person, level, targetLevel int, maternalOrder bool) (string, error) {
	if level > targetLevel {
		return "", nil
	}
	mom, err := p.GetMom()
	if err != nil {
		return "", err
	}
	dad, err := p.GetDad()
	if err != nil {
		return "", err
	}
	var parents []generations.Person
	if maternalOrder {
		parents = []generations.Person{mom, dad}
	} else {
		parents = []generations.Person{dad, mom}
	}

	var sb strings.Builder
	for _, parent := range parents {
		if parent == nil || parent.IsDummy() {
			continue
		}
		sub, err := positionParentSubtree(parent, level+1, targetLevel, maternalOrder)
		if err != nil {
			return "", err
		}
		if sub == "" {
			sb.WriteString(positionNode("p", generations.GenderUnknown, level, targetLevel))
		} else {
			sb.WriteString("parent{\n")
			sb.WriteString(positionNode("g", generations.GenderUnknown, level, targetLevel))
			sb.WriteString(sub)
			sb.WriteString("}\n")
		}
	}
	return sb.String(), nil
}

// positionNode renders a single anonymous node of the given genealogytree
// node type (g or p), marking it if it sits on the target level. Gender is
// passed as a box style because the tiny circles template deletes node
// content, so a sex database field would never be evaluated.
func positionNode(nodeType string, gender generations.Gender, level, targetLevel int) string {
	var boxStyles []string
	if level == targetLevel {
		boxStyles = append(boxStyles, "positionmark")
	} else if level == 0 {
		boxStyles = append(boxStyles, "positionbase")
	} else {
		boxStyles = append(boxStyles, "positionintermediate")
	}
	switch gender {
	case generations.GenderMale:
		boxStyles = append(boxStyles, "positionmale")
	case generations.GenderFemale:
		boxStyles = append(boxStyles, "positionfemale")
	}
	options := ""
	if len(boxStyles) > 0 {
		options = "[box={" + strings.Join(boxStyles, ",") + "}]"
	}
	return fmt.Sprintf("%s%s{}\n", nodeType, options)
}

// commonGender returns the gender shared by all given persons, or
// GenderUnknown if they differ or are unknown.
func commonGender(persons []generations.Person) generations.Gender {
	gender := generations.GenderUnknown
	for i, p := range persons {
		g := p.GetGender()
		if i == 0 {
			gender = g
			continue
		}
		if g != gender {
			return generations.GenderUnknown
		}
	}
	return gender
}
