package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameIsEmpty(t *testing.T) {
	assert.True(t, Name{}.IsEmpty())
	assert.False(t, Name{Last: "Gauss"}.IsEmpty())
	assert.False(t, Name{First: []string{"Carl"}}.IsEmpty())
	assert.False(t, Name{Nick: "C."}.IsEmpty())
}

func TestNameEmpty(t *testing.T) {
	assert.True(t, Name{}.Empty())
	assert.True(t, Name{Nick: "C."}.Empty()) // Empty only checks First, Last, and Birth
	assert.False(t, Name{Last: "Gauss"}.Empty())
	assert.False(t, Name{Birth: "Meier"}.Empty())
}

func TestGetUsedFirst(t *testing.T) {
	// no first names at all
	assert.Equal(t, "", Name{}.GetUsedFirst())

	// first of the first names by default
	n := Name{First: []string{"Carl", "Friedrich"}}
	assert.Equal(t, "Carl", n.GetUsedFirst())

	// explicitly used name wins
	n.Used = "Friedrich"
	assert.Equal(t, "Friedrich", n.GetUsedFirst())

	// unknown used name: all first names joined
	n.Used = "?"
	assert.Equal(t, "Carl Friedrich", n.GetUsedFirst())
}

func TestFormatFull(t *testing.T) {
	n := Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss"}
	assert.Equal(t, "Carl Friedrich Gauss", n.FormatFull())

	n.Birth = "Meier"
	assert.Equal(t, "Carl Friedrich Gauss, geb. Meier", n.FormatFull())
	assert.Equal(t, "Carl Friedrich Gauss", n.FormatFullNoBirth())
}

func TestFormatFullNoMiddle(t *testing.T) {
	n := Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss", Birth: "Meier"}
	assert.Equal(t, "Carl Gauss, geb. Meier", n.FormatFullNoMiddle())
	assert.Equal(t, "Carl Gauss", n.FormatFullNoMiddleNoBirth())

	n.Used = "Friedrich"
	assert.Equal(t, "Friedrich Gauss", n.FormatFullNoMiddleNoBirth())
}

func TestFormatFullInverse(t *testing.T) {
	n := Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss"}
	assert.Equal(t, "Gauss, Carl Friedrich", n.FormatFullInverse())
}
