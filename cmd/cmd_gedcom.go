package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/dimchansky/utfbom"
	gg "github.com/iand/gedcom"
	"github.com/jojomi/generations"
	"github.com/jojomi/generations/gedcom"
	"github.com/jojomi/strtpl"
	"github.com/spf13/cobra"
)

func getGedcomCommand() *cobra.Command {
	var gedcomCmd = cobra.Command{
		Use: "gedcom",
	}
	var gedcomImportCmd = cobra.Command{
		Use: "import",
		Run: gedcomImportHandler,
	}

	var gedcomExportCmd = cobra.Command{
		Use:  "export",
		Args: cobra.MinimumNArgs(1),
		Run:  gedcomExportHandler,
	}

	gedcomCmd.AddCommand(&gedcomImportCmd, &gedcomExportCmd)

	return &gedcomCmd
}

var genderMap = map[generations.Gender]string{
	generations.GenderFemale: "F",
	generations.GenderMale:   "M",
}

func makeIndividualName(p generations.Person) *gedcom.PersonalNameStructure {
	pName := p.GetName()

	personal := strtpl.MustEval(
		`{{- range .First }}{{ . }} {{ end -}}
		{{- if .Used -}} " {{- .Used -}} " {{ end -}}
		{{- if .Nick -}} (" {{- .Nick -}} ") {{ end -}}
		{{- if .Last -}}/{{ .Last }}/ {{- end -}}
		{{- if .Birth -}}, geb. {{ .Birth -}} {{- end -}}
		`,
		pName,
	)

	iName := gedcom.PersonalNameStructure{
		PersonalName: personal,
	}
	return &iName
}

func formatDateGedcom(input string) string {
	if input == "" {
		return ""
	}
	date, err := time.Parse("2006", input)
	if err == nil {
		return input
	}
	date, err = time.Parse("2006-01-02", input)
	if err != nil {
		return "(" + input + ")"
	}
	return strings.ToUpper(date.Format("02 Jan 2006"))
}

func makeIndividual(p generations.Person) *gedcom.IndividualRecord {
	i := gedcom.NewIndividualRecord()

	// XREF
	i.SetXREF(p.GetBestID())
	if v, ok := genderMap[p.GetGender()]; ok {
		i.SetSex(v)
	}

	// Name
	if name := makeIndividualName(p); name != nil {
		i.SetName(name)
	}

	// Events (Birth, Death)
	if event := p.GetBirth(); !event.Empty() {
		i.Events.Birth.Date.Value = formatDateGedcom(event.Date)
		i.Events.Birth.Place.Name = event.Place
	}
	if event := p.GetDeath(); !event.Empty() {
		i.Events.Death.Date.Value = formatDateGedcom(event.Date)
		i.Events.Death.Place.Name = event.Place
	}

	// FAMC
	dad, _ := p.GetDad()
	mom, _ := p.GetMom()
	if key := makeFamilyKey(dad, mom); key != "" {
		i.ChildToFamilyLink = &gedcom.ChildToFamilyLink{
			XREF: key,
		}
	}

	// FAMS
	for _, rel := range p.GetRelationships() {
		key := makeFamilyKey(p, rel.GetPartnerOrDummy())
		i.SpouseToFamilyLink = append(i.SpouseToFamilyLink, &gedcom.SpouseToFamilyLink{
			XREF: key,
		})
	}

	// non-gedcom data
	if childNumber := p.GetChildNumber(); childNumber != nil {
		i.ChildNumber = int(*childNumber)
	}

	return i
}

// makeFamilyKey generates a unique key, male first
func makeFamilyKey(p1, p2 generations.Person) string {
	var orderedPersons []generations.Person
	if p1.GetGender() == generations.GenderMale || p2.GetGender() == generations.GenderFemale {
		orderedPersons = []generations.Person{p1, p2}
	} else {
		orderedPersons = []generations.Person{p2, p1}
	}

	var result string
	for _, p := range orderedPersons {
		if p.IsDummy() {
			continue
		}
		result += p.GetBestID()
	}
	return result
}

func addFamilies(familyMap map[string]*gedcom.FamilyRecord, p generations.Person, indi *gedcom.IndividualRecord) {
	// which family is relevant? mom + dad + person
	dad, _ := p.GetDad()
	mom, _ := p.GetMom()
	key := makeFamilyKey(dad, mom)
	if key == "" {
		return
	}

	// search family
	f, hasFamily := familyMap[key]

	// if family does not exist, create it
	if !hasFamily {
		f = gedcom.NewFamilyRecord()
		f.XREF = key
		if !dad.IsDummy() {
			f.Husband = dad.GetBestID()
		}
		if !mom.IsDummy() {
			f.Wife = mom.GetBestID()
		}
		familyMap[key] = f
	}

	// add child to family
	f.AddChild(indi)

	// add events records (marriage)
	fmt.Println("mom", mom)
	if mom.IsDummy() {
		return
	}
	rel := mom.GetRelationshipWith(dad)
	fmt.Println("rel", rel)
	if rel == nil {
		return
	}
	events := &gedcom.FamilyEventStructure{}
	// TODO marriage place! events.Marriage.Place.Name = rel.
	marriage := rel.GetMarriageReligious()
	if marriage == nil || marriage.Empty() {
		return
	}
	events.Marriage.Date.Value = formatDateGedcom(marriage.Date)
	f.FamilyEventStructure = events
}

func gedcomImportHandler(c *cobra.Command, args []string) {
	data, err := ioutil.ReadFile("example.ged")
	if err != nil {
		panic(err)
	}

	r, _ := utfbom.Skip(bytes.NewReader(data))
	d := gg.NewDecoder(r)

	g, err := d.Decode()
	if err != nil {
		panic(err)
	}

	for _, rec := range g.Individual {
		if len(rec.Name) > 0 {
			println(rec.Name[0].Name)
		}
	}
}

func gedcomExportHandler(c *cobra.Command, args []string) {
	memDB := loadDatabases([]string{
		"herkur-vorfahren-stahl-mit-kindern.yml",
		/// "herkur.yml",
		"Sippe Stahl Stand 1711 Mittag.yml",
	}, "~/.generations/databases/")

	db := gedcom.NewDatabase()
	familyMap := make(map[string]*gedcom.FamilyRecord, len(memDB.Persons)/3)

	// add all individuals and keep track of family associations
	for _, p := range memDB.Persons {
		i := makeIndividual(p)
		db.AddIndividual(i)
		addFamilies(familyMap, p, i)
	}

	// add all families after they are complete
	for _, f := range familyMap {
		db.AddFamily(f)
	}

	fmt.Println(db)
	err := ioutil.WriteFile("test.ged", []byte(db.String()), 0644)
	if err != nil {
		panic(err)
	}

	// doc := gedcom.NewDocument()
	// i := doc.AddIndividual("Testperson")
	// i.AddName("Peter Stahl")
	// i.AddBirthDate("1688-08-17")

	// data := doc.String()
	// fmt.Println(data)

	// buf := bytes.NewBufferString("")
	// encoder := gedcom.NewEncoder(buf, doc)
	// err := encoder.Encode()
	// if err != nil {
	// 	panic(err)
	// }
}
