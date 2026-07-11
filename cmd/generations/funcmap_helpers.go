package main

import (
	htmlTemplate "html/template"
	"strings"
	textTemplate "text/template"
)

func toHTMLFuncMap(funcMap textTemplate.FuncMap) htmlTemplate.FuncMap {
	return htmlTemplate.FuncMap(funcMap)
}

func toTextFuncMap(funcMap htmlTemplate.FuncMap) textTemplate.FuncMap {
	return textTemplate.FuncMap(funcMap)
}

func makeFuncMap(funcMaps ...textTemplate.FuncMap) textTemplate.FuncMap {
	result := textTemplate.FuncMap{}
	for _, funcMap := range funcMaps {
		for key, fun := range funcMap {
			result[key] = fun
		}
	}
	return result
}

func makeHTMLFuncMap(funcMaps ...htmlTemplate.FuncMap) htmlTemplate.FuncMap {
	result := htmlTemplate.FuncMap{}
	for _, funcMap := range funcMaps {
		for key, fun := range funcMap {
			result[key] = fun
		}
	}
	return result
}

func spacingHelpers() textTemplate.FuncMap {
	return textTemplate.FuncMap{
		"space": func(count ...int) string {
			// allows for "space" and "space 4"
			if len(count) == 0 {
				return " "
			}
			return strings.Repeat(" ", count[0])
		},
		"newline": func(count ...int) string {
			// allows for "newline" and "newline 4"
			if len(count) == 0 {
				return "\n"
			}
			return strings.Repeat("\n", count[0])
		},
	}
}

func spacingHelpersHTML() htmlTemplate.FuncMap {
	return htmlTemplate.FuncMap(spacingHelpers())
}
