package generations

import (
	"strings"
)

type Name struct {
	Title string   `yaml:",omitempty"`
	First []string `yaml:",omitempty"`
	// If the used first name is different from the first element in the .First slice, it can be set using .Used
	Used  string `yaml:",omitempty"`
	Last  string `yaml:",omitempty"`
	Birth string `yaml:",omitempty"`
	Alias string `yaml:",omitempty"`
	Nick  string `yaml:",omitempty"`

	Sources ReferenceList `yaml:",omitempty"`
}

func (n Name) IsEmpty() bool {
	return n.Title == "" && len(n.First) == 0 && n.Used == "" && n.Last == "" && n.Birth == "" && n.Alias == "" && n.Nick == ""
}

func (n Name) writeBirthSuffix(b *strings.Builder) {
	if n.Birth != "" {
		b.WriteString(", geb. ")
		b.WriteString(n.Birth)
	}
}

// GetUsedFirst returns the used name if it is given, the first of the first names otherwise
func (n Name) GetUsedFirst() string {
	if n.Used == "?" {
		return strings.Join(n.First, " ")
	}
	if n.Used != "" {
		return n.Used
	}
	if len(n.First) == 0 {
		return ""
	}
	return n.First[0]
}

// FormatFullInverse displays the full name, lastname before firstname
func (n Name) FormatFullInverse() string {
	var b strings.Builder
	if n.Last != "" {
		b.WriteString(n.Last)
		b.WriteString(",")
	}
	for _, first := range n.First {
		b.WriteString(" ")
		b.WriteString(first)
	}
	if n.Used != "" {
		b.WriteString(` "`)
		b.WriteString(n.Used)
		b.WriteString(`"`)
	}
	if n.Nick != "" {
		b.WriteString(` ("`)
		b.WriteString(n.Nick)
		b.WriteString(`")`)
	}
	n.writeBirthSuffix(&b)
	return b.String()
}

// FormatFull displays the full name, firstname before lastname
func (n Name) FormatFull() string {
	var b strings.Builder
	n.writeFullNoBirth(&b)
	n.writeBirthSuffix(&b)
	return b.String()
}

// FormatFullNoBirth displays the full name, firstname before lastname, excluding birth name
func (n Name) FormatFullNoBirth() string {
	var b strings.Builder
	n.writeFullNoBirth(&b)
	return b.String()
}

func (n Name) writeFullNoBirth(b *strings.Builder) {
	for _, first := range n.First {
		b.WriteString(first)
		b.WriteString(" ")
	}
	if n.Used != "" {
		b.WriteString(`"`)
		b.WriteString(n.Used)
		b.WriteString(`" `)
	}
	if n.Nick != "" {
		b.WriteString(`("`)
		b.WriteString(n.Nick)
		b.WriteString(`") `)
	}
	if n.Last != "" {
		b.WriteString(n.Last)
	}
}

// FormatFullNoMiddle displays the full name, firstname before lastname, without middlename(s)
func (n Name) FormatFullNoMiddle() string {
	var b strings.Builder
	n.writeFullNoMiddleNoBirth(&b)
	n.writeBirthSuffix(&b)
	return b.String()
}

// FormatFullNoMiddleNoBirth displays the full name, firstname before lastname, without middlename(s), excluding birth name
func (n Name) FormatFullNoMiddleNoBirth() string {
	var b strings.Builder
	n.writeFullNoMiddleNoBirth(&b)
	return b.String()
}

func (n Name) writeFullNoMiddleNoBirth(b *strings.Builder) {
	if usedFirst := n.GetUsedFirst(); usedFirst != "" {
		b.WriteString(usedFirst)
		b.WriteString(" ")
	}
	if n.Nick != "" {
		b.WriteString(`("`)
		b.WriteString(n.Nick)
		b.WriteString(`") `)
	}
	if n.Last != "" {
		b.WriteString(n.Last)
	}
}
