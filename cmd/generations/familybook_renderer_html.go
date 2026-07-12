package main

import (
	"fmt"
	"html/template"
	htmlTemplate "html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/jojomi/generations"
	"github.com/jojomi/go-script/v2"
	"github.com/jojomi/strtpl"
	"github.com/juju/errors"
	"github.com/spf13/pflag"
)

type FamilybookRendererHTML struct {
	config FamilybookConfig

	targetDir string
}

func NewFamilybookRendererHTML(config FamilybookConfig) *FamilybookRendererHTML {
	return &FamilybookRendererHTML{
		config: config,
	}
}

func (f *FamilybookRendererHTML) Render(outputFilename, templateFilename string, renderConfig RenderConfig) error {
	var (
		err error
		sc  = script.NewContext()
	)

	// prepare target folder
	f.targetDir = sc.AbsPath(filepath.Dir(outputFilename))
	if sc.DirExists(f.targetDir) {
		err = os.RemoveAll(f.targetDir)
		if err != nil {
			return errors.Annotatef(err, "removing existing output directory %s failed", f.targetDir)
		}
	}

	// make dirs
	for _, dir := range []string{f.targetDir, filepath.Join(f.targetDir, "person")} {
		err = sc.EnsureDirExists(dir, 0750)
		if err != nil {
			return errors.Annotatef(err, "could not create output directory %s", dir)
		}
	}

	// loop through different subtemplates (index, persons, and so on)
	tplFilenames := []string{"index.html", "spitzahnen.html", "zeitleiste.html", "recherche.html"}

	for _, tplFilename := range tplFilenames {
		err = f.renderTemplateFile(tplFilename, tplFilename, makeHTMLFuncMap(toHTMLFuncMap(makeCustomFuncMap(nil)), templateHelpers(), toHTMLFuncMap(spacingHelpers()), toHTMLFuncMap(symbols()), personHelpers()), renderConfig)
		if err != nil {
			return errors.Annotatef(err, "rendering template %s failed", tplFilename)
		}
	}

	tplFilename := "person/details.html"
	workerCount := runtime.NumCPU()
	wp := workerpool.New(workerCount)
	errc := make(chan error, 1)
	allPersons := allPersons(renderConfig.Generations, renderConfig.ExtraPersons)
	for _, person := range allPersons.GetPersons() {
		wp.Submit(func(person generations.Person, errorChannel chan error) func() {
			return func() {
				if len(errorChannel) > 0 {
					fmt.Print("-")
					return
				}
				outputFilename := "person/" + person.GetMainTreePositionID() + ".html"
				err = f.renderPerson(tplFilename, outputFilename, renderConfig, makeHTMLFuncMap(templateHelpers(), toHTMLFuncMap(spacingHelpers()), toHTMLFuncMap(symbols()), personHelpers()), allPersons, person)
				if err != nil {
					go func() {
						errc <- errors.Annotatef(err, "rendering template %s to %s for %s failed", tplFilename, outputFilename, person)
					}()
					fmt.Print("F")
				} else {
					fmt.Print(".")
				}
			}
		}(person, errc))
	}
	wp.StopWait()
	fmt.Println()

	if len(errc) > 0 {
		err = <-errc
		fmt.Println(err)
		os.Exit(1)
		return err
	}

	// tailwind
	sc.MustCommandExist("npx")
	productionBuild := true
	if productionBuild {
		sc.SetEnv("NODE_ENV", "production")
	}

	c := script.LocalCommandFrom("npx tailwindcss")
	inputCSS := sc.AbsPath(filepath.Join(f.getTemplateBasePath(), "s.css"))
	outputCSS := sc.AbsPath(filepath.Join(f.targetDir, "dist", "s.css"))
	outputCSSDir := filepath.Dir(outputCSS)
	err = sc.EnsureDirExists(outputCSSDir, 0750)
	if err != nil {
		return errors.Annotatef(err, "could not create output directory %s", outputCSSDir)
	}
	c.AddAll("--config", sc.AbsPath(filepath.Join(f.getTemplateBasePath(), "..", "tailwind.config.js")), "--input", inputCSS, "--output", outputCSS, "--jit")
	if productionBuild {
		c.AddAll("--purge", filepath.Join(f.targetDir, "**/*.html"), "--minify")
	}
	/*fmt.Println(c)
	os.Exit(1)*/

	pr, err := sc.ExecuteDebug(c)
	if err != nil {
		return errors.Annotate(err, "tailwind failed")
	}
	if !pr.Successful() {
		return fmt.Errorf("tailwind failed: %s", pr.Output())
	}

	return nil
}

func (f *FamilybookRendererHTML) GetIntermediateFilename() string {
	return ""
}

func (f *FamilybookRendererHTML) GetOutputFilename() string {
	filename := f.config.OutputFilename

	// ensure suffix
	if !strings.HasSuffix(filename, "index.html") {
		filename = filepath.Join(filename, "index.html")
	}

	return filepath.Join("output", filename)
}

func (f *FamilybookRendererHTML) DoesCompile() bool {
	return false
}

func (f *FamilybookRendererHTML) Compile(flags *pflag.FlagSet, stepFunc func(step string)) error {
	// does not compile
	return nil
}

func (f *FamilybookRendererHTML) getTemplateBasePath() string {
	return filepath.Join("templates", "familybook", "html", "src")
}

func (f *FamilybookRendererHTML) renderTemplateFile(templatePath, relativeOutputPath string, funcMap template.FuncMap, data interface{}) error {
	// read file
	fullTemplatePath := filepath.Join(f.getTemplateBasePath(), templatePath)
	tplData, err := os.ReadFile(fullTemplatePath)
	if err != nil {
		return errors.Annotatef(err, "could not read template %s", fullTemplatePath)
	}

	output, err := strtpl.EvalHTMLWithFuncMap(string(tplData), funcMap, data)
	if err != nil {
		return errors.Annotatef(err, "rendering template %s failed", templatePath)
	}

	// write output
	fullOutputPath := filepath.Join(f.targetDir, relativeOutputPath)
	err = os.WriteFile(fullOutputPath, []byte(output), 0640)
	if err != nil {
		return errors.Annotatef(err, "could not write template output to file %s", fullOutputPath)
	}

	return nil
}

func templateHelpers() template.FuncMap {
	return template.FuncMap{
		"sub":                  subHTML,
		"allPersons":           allPersons,
		"sortByBirthOrBaptism": sortByBirthOrBaptism,
		"sortByDeath":          sortByDeath,
	}
}

func personHelpers() template.FuncMap {
	return template.FuncMap{
		"getPersonDetailsURL": func(person generations.Person) string {
			urlID := ""
			if person.GetMainTreePositionID() != "" {
				urlID = person.GetMainTreePositionID()
			} else {
				urlID = person.GetBestID()
			}
			return "person/" + urlID + ".html"
		},
	}
}

func symbols() textTemplate.FuncMap {
	return textTemplate.FuncMap{
		"symbol": func(name string) (string, error) {
			switch name {
			case "male":
				// 'MALE SIGN' (U+2642)
				return "♂", nil
			case "female":
				// 'FEMALE SIGN' (U+2640)
				return "♀", nil
			case "birth":
				// 'ASTERISK' (U+002A)
				return "*", nil
			case "baptism":
				// 'ALMOST EQUAL TO' (U+2248)
				return "≈", nil
			case "death":
				// 'LATIN CROSS' (U+271D)
				return "✝︎", nil
			case "burial":
				// 'COFFIN' (U+26B0)
				return "⚰︎", nil
			}
			return "", fmt.Errorf("symbol not found: %s", name)
		},
	}
}

func symbolsHTML() htmlTemplate.FuncMap {
	return htmlTemplate.FuncMap(symbols())
}

func (f *FamilybookRendererHTML) renderPerson(
	templatePath, relativeOutputPath string,
	renderConfig RenderConfig,
	funcMap template.FuncMap,
	allPersons *generations.PersonList,
	person generations.Person) error {

	// return nil // TEMP speedup
	referencesContainer := generations.NewReferencesContainer()
	referencesContainer.SetLabelBase(fmt.Sprintf("src:%s-", person.GetBestID()))
	mainSources := person.GetSources()
	for _, m := range mainSources {
		referencesContainer.Add(m, true)
	}

	data := struct {
		Config              *FamilybookConfig
		Person              generations.Person
		Lineages            *generations.LineageContainer
		ZeroDate            time.Time
		AllPersons          *generations.PersonList
		ReferencesContainer *generations.ReferencesContainer
	}{
		Config:              renderConfig.Config,
		Person:              person,
		Lineages:            renderConfig.Lineages,
		ZeroDate:            time.Time{},
		AllPersons:          allPersons,
		ReferencesContainer: referencesContainer,
	}

	customFuncMap := makeCustomFuncMap(map[string]interface{}{
		"findTreePositionID": findTreePositionID,
		"derefInt":           derefInt,
		"derefUint":          derefUint,
		"uint2int":           uint2int,
		"refName":            refName,
		"placeConnection":    placeConnection,
		"printDuration":      printDuration,
	})
	customFuncMap["sub"] = subHTML
	customFuncMap["printDuration"] = printDurationHTML
	allFuncMap := makeHTMLFuncMap(toHTMLFuncMap(customFuncMap), templateHelpers(), toHTMLFuncMap(spacingHelpers()), toHTMLFuncMap(symbols()), personHelpers(), funcMap)

	err := f.renderTemplateFile(templatePath, relativeOutputPath, allFuncMap, data)
	if err != nil {
		return errors.Annotatef(err, "rendering person failed: %v", person)
	}
	return nil
}
