package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jojomi/generations"
	"github.com/stretchr/testify/assert"
)

func positionTestDatabase(t *testing.T) *generations.MemoryDatabase {
	t.Helper()
	data := `persons:
- id: proband
  gender: male
  mom: mama
  dad: papa
- id: mama
  gender: female
  mom: oma
- id: papa
  gender: male
- id: oma
  gender: female
`
	filename := filepath.Join(t.TempDir(), "position.yml")
	err := os.WriteFile(filename, []byte(data), 0644)
	assert.Nil(t, err)

	database := generations.NewMemoryDatabase()
	err = database.ParseYamlFile(filename)
	assert.Nil(t, err)
	return database
}

func TestPositionParentTree(t *testing.T) {
	database := positionTestDatabase(t)
	proband, err := database.GetByID("proband")
	assert.Nil(t, err)

	result, err := positionParentTree([]generations.Person{proband}, 2, true)
	assert.Nil(t, err)
	expected := `parent{
g[box={positionbase,positionmale}]{}
parent{
g[box={positionintermediate}]{}
p[box={positionmark}]{}
}
p[box={positionintermediate}]{}
}
`
	assert.Equal(t, expected, result)
}

func TestPositionParentTreeFirstLevel(t *testing.T) {
	database := positionTestDatabase(t)
	proband, err := database.GetByID("proband")
	assert.Nil(t, err)

	// level 1: parents marked, grandparents pruned, paternal order
	result, err := positionParentTree([]generations.Person{proband}, 1, false)
	assert.Nil(t, err)
	expected := `parent{
g[box={positionbase,positionmale}]{}
p[box={positionmark}]{}
p[box={positionmark}]{}
}
`
	assert.Equal(t, expected, result)
}

func TestPositionParentTreeMixedGenders(t *testing.T) {
	database := positionTestDatabase(t)
	proband, err := database.GetByID("proband")
	assert.Nil(t, err)
	mama, err := database.GetByID("mama")
	assert.Nil(t, err)

	// mixed genders at the root: neutral start node with positionbase
	result, err := positionParentTree([]generations.Person{proband, mama}, 1, true)
	assert.Nil(t, err)
	assert.Contains(t, result, "parent{\ng[box={positionbase}]{}\n")
}

func TestPositionParentTreeNoParents(t *testing.T) {
	database := positionTestDatabase(t)
	oma, err := database.GetByID("oma")
	assert.Nil(t, err)

	result, err := positionParentTree([]generations.Person{oma}, 1, true)
	assert.Nil(t, err)
	assert.Equal(t, "", result)
}
