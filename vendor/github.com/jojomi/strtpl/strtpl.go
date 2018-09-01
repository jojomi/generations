package strtpl

import (
	"bytes"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

// Eval applies Golang's text templating functions on a string with given data and returns the resulting string.
func Eval(templateString string, data interface{}) (output string, err error) {
	return evalTextTemplate(templateString, textTemplate.FuncMap{}, data)
}

// MustEval applies Golang's text templating functions on a string with given data and returns the resulting string.
// In case of errors on the way, this function panics.
func MustEval(templateString string, data interface{}) (output string) {
	return must(Eval(templateString, data))
}

// EvalWithFuncMap allows combining Eval with a custom FuncMap.
func EvalWithFuncMap(templateString string, funcs textTemplate.FuncMap, data interface{}) (output string, err error) {
	return evalTextTemplate(templateString, funcs, data)
}

// MustEvalWithFuncMap allows combining MustEval with a custom FuncMap.
func MustEvalWithFuncMap(templateString string, funcs textTemplate.FuncMap, data interface{}) (output string) {
	return must(EvalWithFuncMap(templateString, funcs, data))
}

// EvalHTML applies Golang's html templating functions on a string with given data and returns the resulting string.
func EvalHTML(templateString string, data interface{}) (output string, err error) {
	return evalHTMLTemplate(templateString, htmlTemplate.FuncMap{}, data)
}

// MustEvalHTML applies Golang's html templating functions on a string with given data and returns the resulting string.
// In case of errors on the way, this function panics.
func MustEvalHTML(templateString string, data interface{}) (output string) {
	return must(EvalHTML(templateString, data))
}

// EvalHTMLWithFuncMap allows combining EvalHTML with a custom FuncMap.
func EvalHTMLWithFuncMap(templateString string, funcs htmlTemplate.FuncMap, data interface{}) (output string, err error) {
	return evalHTMLTemplate(templateString, funcs, data)
}

// MustEvalHTMLWithFuncMap allows combining MustEvalHTML with a custom FuncMap.
func MustEvalHTMLWithFuncMap(templateString string, funcs htmlTemplate.FuncMap, data interface{}) (output string) {
	return must(EvalHTMLWithFuncMap(templateString, funcs, data))
}

// helper functions
func must(output string, err error) string {
	if err != nil {
		panic(err)
	}
	return output
}

func evalTextTemplate(templateString string, funcMap textTemplate.FuncMap, data interface{}) (output string, err error) {
	var outputBuffer bytes.Buffer
	t, err := textTemplate.New("tmpl").Funcs(funcMap).Parse(templateString)
	if err != nil {
		return
	}
	err = t.Execute(&outputBuffer, data)
	if err != nil {
		return
	}
	output = outputBuffer.String()
	return
}

func evalHTMLTemplate(templateString string, funcMap htmlTemplate.FuncMap, data interface{}) (output string, err error) {
	var outputBuffer bytes.Buffer
	t, err := htmlTemplate.New("tmpl").Funcs(funcMap).Parse(templateString)
	if err != nil {
		return
	}
	err = t.Execute(&outputBuffer, data)
	if err != nil {
		return
	}
	output = outputBuffer.String()
	return
}
