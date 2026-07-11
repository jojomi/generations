package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/jojomi/strtpl"

	"github.com/jojomi/generations"
	"github.com/juju/errors"
	"github.com/spf13/cobra"
)

var quotedStringRegexp = regexp.MustCompile(`"([^"]*)"`)

func getYearCommand() *cobra.Command {
	cmd := cobra.Command{
		Use: "year",
		Run: yearHandler,
	}
	flags := cmd.PersistentFlags()
	flags.StringSlice("databases", []string{}, "list of databases to load")
	flags.String("output-filename", "", "output filename, ex. year.tex")
	return &cmd
}

type AnniversaryOptions struct {
	IncludeBirthDays    bool
	IncludeDeathDays    bool
	IncludeMarriageDays bool
}

func yearHandler(cmd *cobra.Command, args []string) {
	var anniversaryOptions = AnniversaryOptions{
		IncludeBirthDays:    true,
		IncludeDeathDays:    true,
		IncludeMarriageDays: true,
	}

	database, err := loadDatabasesFromCLIFlags(cmd)
	if err != nil {
		panic(err)
	}

	// select persons
	persons, err := getPersonListFromCLIFlags(cmd, database)
	if err != nil {
		panic(err)
	}
	persons.RemoveDuplicatesByBestID()

	days := getAnniversaries(persons, anniversaryOptions)

	// sort
	sortByDayOfYear(days)

	// output
	tpl := `
{{- range .Days -}}
{{ .Date.Format "02.01.2006" }} – {{ .Description }} – {{ .Person.GetName.FormatFullNoMiddle | latex }}
{{ end -}}`
	templateBytes, err := ioutil.ReadFile("/home/johannes/go/src/github.com/jojomi/generations/templates/year.tex")
	if err != nil {
		panic(err)
	}
	tpl = string(templateBytes)

	data := map[string]interface{}{
		"Title":  "Jahrestage Cousinen, Cousins, Tanten und Onkel",
		"Date":   time.Now(),
		"Days":   days,
		"Sparse": false,
	}
	funcs := template.FuncMap{
		"latex": func(input string) string {
			return quotedStringRegexp.ReplaceAllString(input, `\enquote{$1}`)
		},
	}
	output := strings.TrimSpace(strtpl.MustEvalWithFuncMap(tpl, funcs, data))

	outputFilename, err := cmd.PersistentFlags().GetString("output-filename")
	if err != nil {
		panic(err)
	}
	if outputFilename == "" {
		fmt.Println(output)
		return
	}
	err = ioutil.WriteFile(outputFilename, []byte(output), 0660)
	if err != nil {
		panic(err)
	}
}

func loadDatabasesFromCLIFlags(cmd *cobra.Command) (generations.Database, error) {
	databases, err := cmd.PersistentFlags().GetStringSlice("databases")
	if err != nil {
		return nil, err
	}

	return generations.LoadDatabases(databases, flagRootDatabaseBaseDir)
}

func getPersonListFromCLIFlags(cmd *cobra.Command, database generations.Database) (*generations.PersonList, error) {
	var (
		persons generations.PersonList
	)
	persons = generations.NewPersonList()
	// root, err := database.GetByID("StahJose1865")
	roots := []string{
		"StahMari1921", // Stahl Enkel und Kinder
		"MitlTher1921", // Mitlmeier Enkel und Kinder
	}
	for _, rootID := range roots {
		root, err := database.GetByID(rootID)
		if err != nil {
			return nil, err
		}
		persons.AddPerson(root)
	}

	// Partner
	partners := addPartnersOf(&persons)
	// Tanten/Onkel + Partner
	children := addPartnersOf(childrenOf(persons))
	// Cousinen/Cousins + Partner
	grandchildren := addPartnersOf(grandchildrenOf(persons))
	grandgrandchildren := addPartnersOf(childrenOf(*grandchildren))

	persons.AddList(partners)
	persons.AddList(children)
	persons.AddList(grandchildren)
	persons.AddList(grandgrandchildren)

	return &persons, nil
}

func addPartnersOf(p *generations.PersonList) *generations.PersonList {
	result := p
	result.AddList(partnersOf(*p))
	return result
}

func partnersOf(p generations.PersonList) *generations.PersonList {
	partners, err := p.GetPartners()
	if err != nil {
		panic(err)
	}
	return &partners
}

func grandchildrenOf(p generations.PersonList) *generations.PersonList {
	children, err := p.GetChildren()
	if err != nil {
		panic(err)
	}
	grandchildren, err := children.GetChildren()
	if err != nil {
		panic(err)
	}
	return &grandchildren
}

func childrenOf(p generations.PersonList) *generations.PersonList {
	children, err := p.GetChildren()
	if err != nil {
		panic(err)
	}
	return &children
}

func getAnniversaries(personList *generations.PersonList, anniversaryOptions AnniversaryOptions) []DayDetails {
	const dateFormat = "2006-01-02"
	const longMarriageDescription = false

	var days []DayDetails
	for _, person := range personList.GetPersons() {
		if birth := person.GetBirth(); anniversaryOptions.IncludeBirthDays && !birth.Empty() {
			day := DayDetails{
				Person: person,
			}
			day.Description = "Geburtstag"
			date, err := time.Parse(dateFormat, birth.Date)
			if err != nil {
				fmt.Println(errors.Annotatef(err, "error: invalid marriage date for %s", person.GetName().FormatFullNoMiddle()))
				continue
			}
			day.Date = date
			days = append(days, day)
		}
		if death := person.GetDeath(); anniversaryOptions.IncludeDeathDays && !death.Empty() {
			day := DayDetails{
				Person: person,
			}
			day.Description = "Todestag"
			date, err := time.Parse(dateFormat, death.Date)
			if err != nil {
				fmt.Println(errors.Annotatef(err, "error: invalid/incomplete death date for %s", person.GetName().FormatFullNoMiddle()))
				continue
			}
			day.Date = date
			days = append(days, day)
		}
		if anniversaryOptions.IncludeMarriageDays {
			for _, relationship := range person.GetRelationships() {
				if relationship.GetMarriage().Empty() {
					// not married
					continue
				}
				partner, err := relationship.GetPartner()
				if err != nil {
					fmt.Println(errors.Annotate(err, "error: partner not found in database"))
					continue
				}
				day := DayDetails{
					Person: person,
				}
				if longMarriageDescription {
					day.Description = "Hochzeitstag mit " + partner.GetName().FormatFullNoMiddle()
				} else {
					day.Description = "Hochzeitstag"
				}
				date, err := time.Parse(dateFormat, relationship.GetMarriage().Date)
				if err != nil {
					fmt.Println(errors.Annotatef(err, "error: invalid marriage date for %s", person.GetName().FormatFullNoMiddle()))
					continue
				}
				day.Date = date
				days = append(days, day)
			}
		}
	}

	return days
}

func sortByDate(days []DayDetails) {
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.Format("2006-01-02") < days[j].Date.Format("2006-01-02")
	})
}

func sortByDayOfYear(days []DayDetails) {
	sort.Slice(days, func(i, j int) bool {
		dayI := days[i].Date.Format("01-02")
		dayJ := days[j].Date.Format("01-02")
		if dayI == dayJ {
			return days[i].Date.Format("2006") < days[j].Date.Format("2006")
		}
		return dayI < dayJ
	})
}

type DayDetails struct {
	Date        time.Time
	Description string
	Person      generations.Person
}
