package generations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*func LevelTestCombine(t *testing.T) {
	tests := []struct {
		Input        LevelConfig
		ProbandLevel int
		Output       []AbsoluteLevel
	}{
		{
			Input: LevelConfig{
				Absolute: []AbsoluteLevel{
					AbsoluteLevel{
						Index: 0,
						Color: LevelColor{
							Main: "green",
						},
					},
				},
			},
			ProbandLevel: 0,
			Output: []AbsoluteLevel{
				AbsoluteLevel{
					Index: 0,
					Color: LevelColor{
						Main: "green",
					},
				},
			},
		},
	}

	for _, test := range tests {
		inp := test.Input
		inp.Combine(test.ProbandLevel)
		assert.True(t, inp.Combined.Equals(test.Output))
	}
}*/

func TestLevelColor(t *testing.T) {
	c := &LevelColor{}
	assert.Empty(t, c.Main)
	assert.Empty(t, c.Leaf)
	c2 := &LevelColor{Main: "red", Leaf: "white"}
	assert.NotEmpty(t, c2.Main)
	assert.NotEmpty(t, c2.Leaf)
	c21 := &LevelColor{Main: "orange", Leaf: ""}

	c3 := c.OverwriteWith(c2)
	assert.Equal(t, "red", c3.Main)
	assert.Equal(t, "white", c3.Leaf)
	c4 := c2.OverwriteWith(c)
	assert.Empty(t, c4.Main)
	assert.Empty(t, c4.Leaf)

	// TODO MergeWithBase
	c5 := c21.MergeWithBase(c2)
	assert.Equal(t, "orange", c5.Main)
	assert.Equal(t, "white", c5.Leaf)
	c6 := c2.MergeWithBase(c21)
	assert.Equal(t, "red", c6.Main)
	assert.Equal(t, "white", c6.Leaf)
}

func TestLevelOptions(t *testing.T) {
	o := LevelOptions("")
	o2 := LevelOptions("red")
	assert.Equal(t, "red", string(*o.Merge(o2)))
	assert.Equal(t, "red", string(*o2.Merge(o)))
}

func TestLevelMerge(t *testing.T) {
	l1 := AbsoluteLevel{
		Color: LevelColor{
			Main: "blue",
		},
		Options: "l1-opts",
	}
	l2 := AbsoluteLevel{
		Color: LevelColor{
			Leaf: "pink",
		},
		Options: "l2-opts",
	}
	merged := Merge(&l1, &l2)
	assert.Equal(t, "l2-opts%\n%\nl1-opts", string(merged.GetOptions()))
}

func TestLevelConfigCombine(t *testing.T) {
	lc := LevelConfig{
		Absolute: []AbsoluteLevel{
			AbsoluteLevel{
				Index: 0,
				Color: LevelColor{
					Main: "green",
				},
			},
		},
		Relative: []RelativeLevel{
			RelativeLevel{
				Index: -5,
				Color: LevelColor{
					Main: "red",
				},
			},
		},
	}
	lc.Combine(5)
	assert.Equal(t, "red", lc.Combined[0].Color.Main)
}
