package main

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"image"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jojomi/generations"
	script "github.com/jojomi/go-script"
	"github.com/jojomi/strtpl"
	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
	"github.com/vincent-petithory/dataurl"
)

func getCousinsCommand() *cobra.Command {
	cmd := cobra.Command{
		Use: "cousins",
		Run: commandCousins,
	}
	return &cmd
}

type FBColor struct {
	Foreground string
	Background string
}
type CousinEntry struct {
	Cousin   generations.Person
	Partners generations.PersonList
	Parent   generations.Person
	Color    FBColor
}

type CousinsExportOptions struct {
	UseImages    bool
	UseColors    bool
	SingleColumn bool
}

func commandCousins(c *cobra.Command, args []string) {
	databaseFiles := []string{
		"herkur.yml",
		"Familie Agnes Ohrnberger.yml",
		"Familie Alois Stahl.yml",
		"Familie Anton Stahl.yml",
		"Familie Bock.yml",
		"Familie Dick.yml",
		"Familie Frank.yml",
		"Familie Johann Stahl.yml",
		"Familie Karl Stahl.yml",
		"Familie Monika Kilb.yml",
		"Familie Helene Mark.yml",
		/// "Daten Brigitte.yml",
	}
	// https://www.w3schools.com/colors/colors_names.asp
	colors := map[string]FBColor{
		"StahKarl1897": FBColor{Background: "Coral"},
		"BockAnna1898": FBColor{Background: "DarkKhaki"},
		"BockLudw1893": FBColor{Background: "DarkKhaki"},
		"FranRosa1899": FBColor{Background: "DeepSkyBlue"},
		"StahMart1900": FBColor{Background: ""},
		"StahAnto1901": FBColor{Background: "RoyalBlue", Foreground: "white"},
		"StahJoha1903": FBColor{Background: "HotPink", Foreground: "white"},
		"DickMari1905": FBColor{Background: "DarkRed", Foreground: "white"},
		"MarkHele1907": FBColor{Background: "SeaGreen", Foreground: "white"},
		"StahAloi1909": FBColor{Background: "SpringGreen"},
		"KilbMoni1911": FBColor{Background: "Gold"},
		"StahJose1912": FBColor{Background: "Silver"},
		"StahAugu1914": FBColor{Background: "Tomato", Foreground: "white"},
		"OhrnAgne1916": FBColor{Background: "MediumTurquoise"},
	}

	database, err := generations.LoadDatabases(databaseFiles, flagRootDatabaseBaseDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p, err := database.GetByID("StahJose1865") // Josef Felix Stahl (wmB5aqL52M)
	// p, err := database.GetByID("BrümLore1879")
	// p, err := database.GetByID("StahMari1921") // Maria Stahl
	// p, err := database.GetByID("KlimElis1929")
	// p, err := database.GetByID("MlynMart1923")
	if err != nil {
		panic(err)
	}
	grandChildren, err := getGrandChildren(p, colors)
	if err != nil {
		panic(err)
	}
	// add children of Anna Bock (Halbcousins)
	/*ludwigBock, err := database.GetByID("BockLudw1893")
	if err != nil {
		panic(err)
	}
	halbcousins, err := ludwigBock.GetChildren()
	if err != nil {
		panic(err)
	}
	color, _ := colors[ludwigBock.GetID()]
	for _, halbcousin := range halbcousins {
		ce := getCousinEntry(halbcousin, ludwigBock, color)
		grandChildren = append(grandChildren, ce)
	}*/

	// sort
	sort.Slice(grandChildren, func(i, j int) bool {
		a := grandChildren[i].Cousin.GetBirth().Date
		b := grandChildren[j].Cousin.GetBirth().Date
		// a := grandChildren[i].GetBirth().Date[4:]
		// b := grandChildren[j].GetBirth().Date[4:]
		return strings.Compare(a, b) < 0
	})

	// remove duplicates
	grandChildrenFiltered := make([]CousinEntry, 0, len(grandChildren))
	var (
		lastEntry *CousinEntry
	)
	for _, grandChild := range grandChildren {
		if lastEntry != nil && lastEntry.Cousin.GetID() == grandChild.Cousin.GetID() {
			if image := grandChild.Cousin.GetImageFilename(); image != "" && image != lastEntry.Cousin.GetImageFilename() {
				lastEntry.Cousin.SetImageFilename(image)
			}
			continue
		}
		grandChildrenFiltered = append(grandChildrenFiltered, grandChild)
		x := grandChild
		lastEntry = &x
	}

	// print
	var (
		deathMarker string
		deathDate   string
	)
	for _, grandChild := range grandChildrenFiltered {
		deathMarker = ""
		if deathDate = grandChild.Cousin.GetDeath().Date; deathDate != "" {
			deathMarker = ", † " + deathDate
		}
		fmt.Printf("%s: %s%s\n", grandChild.Cousin.GetBirth().Date, formatName(grandChild.Cousin.GetName()), deathMarker)
	}

	// render
	tplContent, err := ioutil.ReadFile("templates/cousins.html")
	if err != nil {
		panic(err)
	}
	exportOptions := CousinsExportOptions{
		UseImages:    false,
		UseColors:    true,
		SingleColumn: true,
	}
	data := map[string]interface{}{
		"Data":    grandChildrenFiltered,
		"Options": exportOptions,
	}
	functions := htmlTemplate.FuncMap{
		"parseDate": func(input string) (time.Time, error) {
			return time.Parse("2006-01-02", input)
		},
		"renderDate": func(input string) string {
			date, err := time.Parse("2006-01-02", input)
			if err == nil {
				return date.Format("02.01.2006")
			}
			return input
		},
		"urlify": func(input string) htmlTemplate.URL {
			return htmlTemplate.URL(input)
		},
		"minus": func(a, b int) int {
			return a - b
		},
		"dataURI": func(imageFilepath string) (htmlTemplate.URL, error) {
			fullPath := filepath.Join("..", "cmd", "output", imageFilepath)
			image, err := getImageFromPath(fullPath)
			if err != nil {
				return htmlTemplate.URL(""), err
			}
			thumbnail := resize.Thumbnail(100, 100, *image, resize.Lanczos2)
			buf := new(bytes.Buffer)
			err = png.Encode(buf, thumbnail)
			if err != nil {
				return htmlTemplate.URL(""), err
			}
			data := buf.Bytes()
			dataURIdata := dataurl.EncodeBytes(data)
			return htmlTemplate.URL(dataURIdata), nil
		},
	}
	tplResult, err := strtpl.EvalHTMLWithFuncMap(string(tplContent), functions, data)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("output/cousins.html", []byte(tplResult), os.FileMode(int(0640)))
	if err != nil {
		panic(err)
	}

}

func getImageFromPath(path string) (*image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	return &m, err
	// if strings.HasSuffix(strings.ToLower(path), ".jpg") {
	// }
	// if strings.HasSuffix(strings.ToLower(path), ".png") {
	// 	m, _, err := image.Decode(reader)
	// 	return &m, err
	// }
	// return nil, nil
}

func formatName(name generations.Name) string {
	return strtpl.MustEval(`{{ index .First 0}} {{ .Last }} {{- if .Birth }}, geb. {{ .Birth }}{{ end }}`, name)
}

func getImage(p generations.Person) string {
	sc := script.NewContext()
	extensions := []string{"jpg", "JPG", "png", "PNG"}
	id := p.GetID()
	for _, ext := range extensions {
		if sc.FileExists(filepath.Join("output", "img", id+"."+ext)) {
			return id + "." + ext
		}
	}
	return ""
}

func getGrandChildren(p generations.Person, colors map[string]FBColor) ([]CousinEntry, error) {
	result := make([]CousinEntry, 0)
	c, err := p.GetChildren()
	if err != nil {
		return result, nil
	}
	var (
		color FBColor
	)
	for _, child := range c.GetPersons() {
		color, _ = colors[child.GetID()]
		gcs, err := child.GetChildren()
		if err != nil {
			return result, nil
		}
		for _, gc := range gcs.GetPersons() {
			ces := getCousinEntries(gc, child, color, true)
			result = append(result, ces...)
		}
	}
	return result, nil
}

func getCousinEntries(person, parent generations.Person, color FBColor, unwrapPartners bool) []CousinEntry {
	result := make([]CousinEntry, 0)

	if image := getImage(person); image != "" {
		person.SetImageFilename(filepath.Join("img", image))
	} else {
		if person.GetImageFilename() != "" {
			person.SetImageFilename(filepath.Join("..", "pictures", person.GetImageFilename()))
		}
	}
	partners, err := person.GetPartners()
	if err != nil {
		panic(err)
	}

	fixedPartners := generations.NewPersonList()
	for _, partner := range partners.GetPersons() {
		if image := getImage(partner); image != "" {
			partner.SetImageFilename(filepath.Join("img", image))
		} else {
			if partner.GetImageFilename() != "" {
				partner.SetImageFilename(filepath.Join("..", "pictures", partner.GetImageFilename()))
			}
		}
		if unwrapPartners {
			result = append(result, CousinEntry{
				Cousin: partner,
				Color:  color,
				Parent: parent,
			})
		} else {
			fixedPartners.AddPerson(partner)
		}
	}

	result = append(result, CousinEntry{
		Cousin:   person,
		Partners: fixedPartners,
		Parent:   parent,
		Color:    color,
	})
	return result
}
