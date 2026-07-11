package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type FamilybookRendererHTML struct {
	config FamilybookConfig
}

func NewFamilybookRendererHTML(config FamilybookConfig) *FamilybookRendererHTML {
	return &FamilybookRendererHTML{
		config: config,
	}
}

func (f *FamilybookRendererHTML) GetIntermediateFilename() string {
	return ""
}

func (f *FamilybookRendererHTML) GetMinifiedFilename() string {
	return ""
}

func (f *FamilybookRendererHTML) GetOutputFilename() string {
	filename := f.config.OutputFilename

	// ensure suffix
	if !strings.HasSuffix(filename, ".html") {
		filename = filename + ".html"
	}

	return filepath.Join("output", filename)
}

func (f *FamilybookRendererHTML) DoesCompile() bool {
	return false
}

func (f *FamilybookRendererHTML) Compile(flags *pflag.FlagSet) error {
	// does not compile
	return nil
}

func (f *FamilybookRendererHTML) DoesMinify() bool {
	return false
}

func (f *FamilybookRendererHTML) Minify() error {
	// does not minify
	return nil
}
