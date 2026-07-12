package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jojomi/generations"
	"github.com/stretchr/testify/assert"
)

func searchIndexTestDatabase(t *testing.T) *generations.MemoryDatabase {
	t.Helper()
	data := `persons:
- id: full
  gender: male
  name:
    first: ["Peter"]
    last: Stahl
  birth:
    date: "1643-03-12"
  death:
    date: "1700-05-20"
  residences:
    - address: "Hauptstr. 5"
      zip-code: "74670"
      city: "Forchtenberg"
  jobs:
    - type: "Bauer"
      company: "Hof Stahl"
  attributes:
    - "Vollspitzahn"
- id: minimal
  gender: female
  name:
    first: ["Anna"]
    last: Muster
`
	filename := filepath.Join(t.TempDir(), "search-index.yml")
	err := os.WriteFile(filename, []byte(data), 0644)
	assert.Nil(t, err)

	database := generations.NewMemoryDatabase()
	err = database.ParseYamlFile(filename)
	assert.Nil(t, err)
	return database
}

func TestBuildSearchIndexFullPerson(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("full")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	entries := buildSearchIndex(&list)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "Peter Stahl", entry.Name)
	assert.Equal(t, "person/full.html", entry.URL)
	assert.Equal(t, []SearchField{
		{Label: "Name", Value: "Peter Stahl"},
		{Label: "Geboren", Value: "12.03.1643"},
		{Label: "Gestorben", Value: "20.05.1700"},
		{Label: "Wohnort", Value: "Hauptstr. 5, 74670 Forchtenberg"},
		{Label: "Beruf", Value: "Bauer (Hof Stahl)"},
		{Label: "Tag", Value: "Vollspitzahn"},
	}, entry.Fields)
}

func TestBuildSearchIndexMinimalPerson(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("minimal")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	entries := buildSearchIndex(&list)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "Anna Muster", entry.Name)
	assert.Equal(t, "person/minimal.html", entry.URL)
	assert.Equal(t, []SearchField{
		{Label: "Name", Value: "Anna Muster"},
	}, entry.Fields)
}

func TestBuildSearchIndexEmptyList(t *testing.T) {
	list := generations.NewPersonList()
	entries := buildSearchIndex(&list)
	assert.Equal(t, []SearchEntry{}, entries)
}

func TestBuildSearchIndexJSON(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("minimal")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	js, err := buildSearchIndexJSON(&list)
	assert.Nil(t, err)

	var decoded []SearchEntry
	err = json.Unmarshal([]byte(js), &decoded)
	assert.Nil(t, err)
	assert.Equal(t, buildSearchIndex(&list), decoded)
}
