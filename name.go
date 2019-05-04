package generations

import "github.com/jojomi/strtpl"

type Name struct {
	Title string   `yaml:"title,omitempty"`
	First []string `yaml:"first,omitempty"`
	// If the used first name is different from the first element in the .First slice, it can be set using .Used
	Used  string `yaml:"used,omitempty"`
	Last  string `yaml:"last,omitempty"`
	Birth string `yaml:"birth,omitempty"`
	Alias string `yaml:"alias,omitempty"`
	Nick  string `yaml:"nick,omitempty"`
}

// GetUsedFirst returns the used name if it is given, the first of the first names otherwise
func (n Name) GetUsedFirst() string {
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
	return strtpl.MustEval(
		`{{- if .Last -}} {{ .Last }}, {{- end -}}
		{{- range .First }} {{ . }}{{- end -}}
		{{- if .Used }} " {{- .Used -}} " {{- end -}}
		{{- if .Nick }} (" {{- .Nick -}} ") {{- end -}}
		{{- if .Birth -}}, geb. {{ .Birth -}} {{- end -}}
		`,
		n,
	)
}

// FormatFull displays the full name, firstname before lastname
func (n Name) FormatFull() string {
	return strtpl.MustEval(
		`{{- range .First }}{{ . }} {{ end -}}
		{{- if .Used -}} " {{- .Used -}} " {{ end -}}
		{{- if .Nick -}} (" {{- .Nick -}} ") {{ end -}}
		{{- if .Last -}}{{ .Last }} {{- end -}}
		{{- if .Birth -}}, geb. {{ .Birth -}} {{- end -}}
		`,
		n,
	)
}

// FormatFullNoMiddle displays the full name, firstname before lastname, withotu middlename(s)
func (n Name) FormatFullNoMiddle() string {
	return strtpl.MustEval(
		`{{- if .GetUsedFirst -}} {{ .GetUsedFirst }} {{ end -}}
		{{- if .Nick -}} (" {{- .Nick -}} ") {{ end -}}
		{{- if .Last -}}{{ .Last }} {{- end -}}
		{{- if .Birth -}}, geb. {{ .Birth -}} {{- end -}}
		`,
		n,
	)
}
