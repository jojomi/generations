package generations

import (
	"strconv"
	"testing"
)

func BenchmarkNameFormatFull(b *testing.B) {
	n := Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss", Birth: "Meier"}
	for i := 0; i < b.N; i++ {
		_ = n.FormatFull()
	}
}

func BenchmarkNameFormatFullNoMiddleNoBirth(b *testing.B) {
	n := Name{First: []string{"Carl", "Friedrich"}, Last: "Gauss"}
	for i := 0; i < b.N; i++ {
		_ = n.FormatFullNoMiddleNoBirth()
	}
}

func BenchmarkPersonString(b *testing.B) {
	p := FlatPerson{ID: "gauss", Name: Name{First: []string{"Carl"}, Last: "Gauss"}}
	for i := 0; i < b.N; i++ {
		_ = p.String()
	}
}

func BenchmarkGetLabelFor(b *testing.B) {
	f := NewFootnoteContainer()
	for i := 0; i < b.N; i++ {
		_ = f.GetLabelFor("Kirchenbuch, Seite 4")
	}
}

func BenchmarkGetFourFourYearID(b *testing.B) {
	p := &FlatPerson{
		Name:  Name{First: []string{"Carl"}, Last: "Gauss"},
		Birth: DatePlace{Date: "1777-04-30"},
	}
	for i := 0; i < b.N; i++ {
		_, _ = GetFourFourYearID(p)
	}
}

func BenchmarkRemoveDuplicatesByBestID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		list := NewPersonList()
		for j := 0; j < 500; j++ {
			list.AddPerson(&FlatPerson{ID: "p" + strconv.Itoa(j%250)})
		}
		b.StartTimer()
		list.RemoveDuplicatesByBestID()
	}
}

func BenchmarkGetWithMergedTreePositionIDs(b *testing.B) {
	list := NewPersonList()
	for j := 0; j < 500; j++ {
		list.AddPerson(&FlatPerson{ID: "p" + strconv.Itoa(j%250)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = list.GetWithMergedTreePositionIDs()
	}
}
