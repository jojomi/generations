package main

import (
	"time"

	"github.com/jojomi/generations"
)

type Template struct {
	Filename string                 `yaml:"filename,omitempty"`
	Options  map[string]interface{} `yaml:"options,omitempty"`
}

type Config struct {
	Databases []string `yaml:"databases,omitempty"`
	Templates struct {
		Document Template `yaml:"document,omitempty"`
		Tree     Template `yaml:"tree,omitempty"`
	} `yaml:"templates,omitempty"`

	DocumentOptions string    `yaml:"document-options,omitempty"`
	Title           string    `yaml:"title,omitempty"`
	Date            time.Time `yaml:"date,omitempty"`
	DateFormat      string    `yaml:"date-format,omitempty"`

	Attribution string                  `yaml:"attribution,omitempty"`
	PreContent  string                  `yaml:"pre-content,omitempty"`
	PostContent string                  `yaml:"post-content,omitempty"`
	CustomHead  string                  `yaml:"custom-head,omitempty"`
	Levels      generations.LevelConfig `yaml:"levels,omitempty"`

	CustomStyles string `yaml:"custom-styles,omitempty"`
	CustomDraw   string `yaml:"custom-draw,omitempty"`

	Trees         []TreeConfig `yaml:"trees"`
	RenderedTrees string       `yaml:"-"`

	OutputFilename string `yaml:"output-filename,omitempty"`
}

type TreeConfig struct {
	Databases []string `yaml:"databases,omitempty"`
	Templates struct {
		Tree Template `yaml:"tree,omitempty"`
	} `yaml:"templates,omitempty"`

	Date       time.Time `yaml:"date,omitempty"`
	DateFormat string    `yaml:"date-format,omitempty"`

	Title       string `yaml:"title,omitempty"`
	Attribution string `yaml:"attribution,omitempty"`

	Proband      string                  `yaml:"proband,omitempty"`
	ProbandLevel int                     `yaml:"proband-level,omitempty"`
	Levels       generations.LevelConfig `yaml:"levels,omitempty"`

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
	if c.Templates.Document.Filename == "" {
		c.Templates.Document.Filename = "templates/document/basic.tex"
	}
	if c.Templates.Tree.Filename == "" {
		c.Templates.Tree.Filename = "templates/tree/basic.tex"
	}
	if c.Date.IsZero() {
		c.Date = time.Now()
	}
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
	if t.Date.IsZero() {
		t.Date = config.Date
	}
	if t.DateFormat == "" {
		t.DateFormat = config.DateFormat
	}
	// Templates
	if t.Templates.Tree.Filename == "" {
		t.Templates.Tree.Filename = config.Templates.Tree.Filename
	}
	if t.Templates.Tree.Options == nil {
		t.Templates.Tree.Options = config.Templates.Tree.Options
	}
}
