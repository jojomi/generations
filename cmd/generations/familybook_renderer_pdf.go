package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type FamilybookRendererPDF struct {
	config FamilybookConfig
}

func NewFamilybookRendererPDF(config FamilybookConfig) *FamilybookRendererPDF {
	return &FamilybookRendererPDF{
		config: config,
	}
}

func (f *FamilybookRendererPDF) GetIntermediateFilename() string {
	return strings.Replace(f.GetOutputFilename(), ".pdf", ".tex", -1)
}

func (f *FamilybookRendererPDF) GetMinifiedFilename() string {
	return strings.Replace(f.GetOutputFilename(), ".pdf", ".small.pdf", -1)
}

func (f *FamilybookRendererPDF) GetOutputFilename() string {
	filename := f.config.OutputFilename

	// ensure suffix
	if !strings.HasSuffix(filename, ".pdf") {
		filename = filename + ".pdf"
	}

	return filepath.Join("output", filename)
}

func (f *FamilybookRendererPDF) DoesCompile() bool {
	return true
}

func (f *FamilybookRendererPDF) Render(outputFilename, templateFilename string, renderConfig RenderConfig) error {
	customFuncs := makeCustomFuncMap(map[string]interface{}{
		"sub":                   sub,
		"renderFamilybookEntry": renderFamilybookEntry,
		"allPersons":            allPersons,
		"percentage":            percentage,
		"percentageInt":         percentageInt,
		"sparseDate":            sparseDate,
		"zeroDate":              func() time.Time { return time.Time{} },
		"derefInt":              derefInt,
		"derefUint":             derefUint,
		"uint2int":              uint2int,
		"refName":               refName,
		"placeConnection":       placeConnection,
		"printDuration":         printDuration,
	})
	return renderDocument(templateFilename, renderConfig, customFuncs, outputFilename)
}

func (f *FamilybookRendererPDF) Compile(flags *pflag.FlagSet, stepFunc func(step string)) error {
	return compileDocument(
		f.GetIntermediateFilename(),
		flagFamilybookNumCompileRuns,
		flagRootVerbose,
		stepFunc,
	)
}

func (f *FamilybookRendererPDF) DoesMinify() bool {
	return true
}

func (f *FamilybookRendererPDF) Minify() error {
	smallFilename := f.GetMinifiedFilename()
	return minifyDocument(
		f.GetOutputFilename(),
		smallFilename,
	)
}
