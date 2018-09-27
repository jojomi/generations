package main

import "github.com/jojomi/generations"

type Config struct {
	Databases []string `yaml:"databases,omitempty"`
	Template  string   `yaml:"template,omitempty"`

	DocumentOptions string `yaml:"document-options,omitempty"`
	Title           string `yaml:"title,omitempty"`
	Date            string `yaml:"date,omitempty"`

	Attribution string `yaml:"attribution,omitempty"`
	PreContent  string `yaml:"pre-content,omitempty"`
	PostContent string `yaml:"post-content,omitempty"`
	CustomHead  string `yaml:"custom-head,omitempty"`

	CustomStyles string `yaml:"custom-styles,omitempty"`
	CustomDraw   string `yaml:"custom-draw,omitempty"`

	Trees []TreeConfig `yaml:"trees"`

	OutputFilename string `yaml:"output-filename,omitempty"`
}

type TreeConfig struct {
	Databases []string `yaml:"databases,omitempty"`

	Title       string `yaml:"title,omitempty"`
	Date        string `yaml:"date,omitempty"`
	Attribution string `yaml:"attribution,omitempty"`

	Proband      string `yaml:"proband,omitempty"`
	ProbandLevel int    `yaml:"proband-level,omitempty"`

	PreContent  string `yaml:"pre-content,omitempty"`
	PostContent string `yaml:"post-content,omitempty"`
	Content     string `yaml:"-,omitempty"`

	CustomStyles string `yaml:"custom-styles,omitempty"`
	CustomDraw   string `yaml:"custom-draw,omitempty"`

	Scale float64 `yaml:"scale,omitempty"`

	PageBreakAfter bool `yaml:"page-break-after,omitempty"`

	RenderTreeOptions generations.RenderTreeOptions `yaml:"render-tree-options"`
}

func (c *Config) SetDefaults() {
}

func (t *TreeConfig) AddGlobals(config Config) {
	if len(t.Databases) == 0 {
		t.Databases = config.Databases
	}
	if t.CustomDraw == "" {
		t.CustomDraw = config.CustomDraw
	}
	if t.CustomStyles == "" {
		t.CustomStyles = config.CustomStyles
	}
}
