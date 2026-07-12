package main

import "github.com/spf13/pflag"

type FamilybookRenderer interface {
	GetIntermediateFilename() string
	GetOutputFilename() string

	Render(outputFilename, templateFilename string, renderConfig RenderConfig) error
	DoesCompile() bool
	Compile(flags *pflag.FlagSet, stepFunc func(step string)) error
}
