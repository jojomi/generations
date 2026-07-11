package main

import "github.com/spf13/pflag"

type FamilybookRenderer interface {
	GetIntermediateFilename() string
	GetMinifiedFilename() string
	GetOutputFilename() string

	DoesCompile() bool
	Compile(flags *pflag.FlagSet) error
	DoesMinify() bool
	Minify() error
}
