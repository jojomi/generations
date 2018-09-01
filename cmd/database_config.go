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

	Trees []TreeConfig `yaml:"trees"`
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
