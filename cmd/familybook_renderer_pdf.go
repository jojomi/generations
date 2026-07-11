package main

import (
	"path/filepath"
	"strings"

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

func (f *FamilybookRendererPDF) Compile(flags *pflag.FlagSet) error {
	return compileDocument(
		f.GetIntermediateFilename(),
		flagFamilybookNumCompileRuns,
		flagRootVerbose,
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
