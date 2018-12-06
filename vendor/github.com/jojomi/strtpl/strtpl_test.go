package strtpl

import (
	"fmt"
	htmlTemplate "html/template"
	"strings"
	"testing"
	textTemplate "text/template"

	"github.com/stretchr/testify/assert"
)

func TestWithoutReplacement(t *testing.T) {
	var tests = []struct {
		input    string
		data     interface{}
		expected string
	}{
		{"standard string", nil, "standard string"},
		{"", false, ""},
		{"my string close to a { .Template }", map[string]string{"Template": "content"}, "my string close to a { .Template }"},
		{"<html>{{ .Content }}</html>", map[string]string{"Content": "<b>test</b>"}, "<html><b>test</b></html>"},
	}

	for _, tt := range tests {
		actual, err := Eval(tt.input, tt.data)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestReplacement(t *testing.T) {
	var tests = []struct {
		input    string
		data     interface{}
		expected string
	}{
		{"replacement from {{ .Map }}", map[string]string{"Map": "any map"}, "replacement from any map"},
		{"replacement from {{ .Struct }} and {{ .Struct }} really", struct{ Struct string }{Struct: "any struct"}, "replacement from any struct and any struct really"},
		{"replacement from string {{ . }}", "this is given here", "replacement from string this is given here"},
	}

	for _, tt := range tests {
		actual, err := Eval(tt.input, tt.data)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestAdvanced(t *testing.T) {
	var tests = []struct {
		input    string
		data     interface{}
		expected string
	}{
		{"Dear {{ if .Male }}Mr.{{ else }}Mrs.{{ end }} {{ .Name -}} !", struct {
			Male bool
			Name string
		}{
			Male: false,
			Name: "Scully",
		}, "Dear Mrs. Scully!"},
		{"Dear {{ if .Male }}Mr.{{ else }}Mrs.{{ end }} {{ .Name -}} !", struct {
			Male bool
			Name string
		}{
			Male: true,
			Name: "Scully",
		}, "Dear Mr. Scully!"},
	}

	for _, tt := range tests {
		actual, err := Eval(tt.input, tt.data)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestAdvancedHTML(t *testing.T) {
	var tests = []struct {
		input      string
		data       interface{}
		expected   string
		expectFail bool
	}{
		{"<html>{{ .Content }}</html>", map[string]string{"Content": "<b>test</b>"}, "<html>&lt;b&gt;test&lt;/b&gt;</html>", false},
		{"No valid template {{ .Data ", nil, "", true},
		{"No valid data {{ .Data }}", struct{}{}, "", true},
	}

	for _, tt := range tests {
		actual, err := EvalHTML(tt.input, tt.data)
		assert.Equal(t, tt.expectFail, err != nil)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestWithInvalidTemplate(t *testing.T) {
	var tests = []struct {
		input string
	}{
		{"Not valid for parser {{ .Data }"},
	}

	for _, tt := range tests {
		_, err := Eval(tt.input, nil)
		assert.NotNil(t, err)
	}
}

func TestWithInvalidData(t *testing.T) {
	var tests = []struct {
		input string
		data  interface{}
	}{
		{"No valid data {{ .Data }}", struct{}{}},
	}

	for _, tt := range tests {
		actual, err := Eval(tt.input, tt.data)
		fmt.Println(actual)
		assert.NotNil(t, err)
	}
}

func TestMustEval(t *testing.T) {
	var tests = []struct {
		input    string
		data     interface{}
		expected string
	}{
		{"no replacement", map[string]string{"Map": "any map"}, "no replacement"},
	}

	for _, tt := range tests {
		actual := MustEval(tt.input, tt.data)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestMustEvalFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	MustEval("abc {{ .invalid", nil)
}

func TestMustEvalHTML(t *testing.T) {
	var tests = []struct {
		input    string
		data     interface{}
		expected string
	}{
		{"no replacement", map[string]string{"Map": "any map"}, "no replacement"},
	}

	for _, tt := range tests {
		actual := MustEvalHTML(tt.input, tt.data)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestMustEvalWithFuncMap(t *testing.T) {
	revFunc := textTemplate.FuncMap{"upper": func(input string) string { return strings.ToUpper(input) }}
	data := map[string]string{
		"fruit": "bananas",
	}

	var tests = []struct {
		input    string
		funcMap  textTemplate.FuncMap
		data     interface{}
		expected string
	}{
		{"no replacement", revFunc, data, "no replacement"},
		{`abc{{ "dEf" | upper }}ghi`, revFunc, data, "abcDEFghi"},
		{`I want {{ upper .fruit }}.`, revFunc, data, "I want BANANAS."},
	}

	for _, tt := range tests {
		actual := MustEvalWithFuncMap(tt.input, tt.funcMap, tt.data)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestMustEvalHTMLWithFuncMap(t *testing.T) {
	revFunc := htmlTemplate.FuncMap{"upper": func(input string) string { return strings.ToUpper(input) }}
	data := map[string]string{
		"fruit": "bananas",
	}

	var tests = []struct {
		input    string
		funcMap  htmlTemplate.FuncMap
		data     interface{}
		expected string
	}{
		{"no replacement", revFunc, data, "no replacement"},
		{`abc{{ "dEf" | upper }}ghi`, revFunc, data, "abcDEFghi"},
		{`I want {{ upper .fruit }}.`, revFunc, data, "I want BANANAS."},
	}

	for _, tt := range tests {
		actual := MustEvalHTMLWithFuncMap(tt.input, tt.funcMap, tt.data)
		assert.Equal(t, tt.expected, actual, "they should be equal")
	}
}

func TestMustEvalHTMLFail(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	MustEvalHTML("abc {{ .invalid", nil)
}
