package main

type OutputFilenameConfig interface {
	GetOutputFilename() string
	SetOutputFilename(filename string)
}
