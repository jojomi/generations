//go:generate gonum -types=SexValue

package gedcom

type SexValue struct {
	Male   string `enum:"M"`
	Female string `enum:"F"`
}
