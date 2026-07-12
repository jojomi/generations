package main

import (
	"fmt"
	htmlTemplate "html/template"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/bearbin/go-age"
	"github.com/jojomi/generations"
	script "github.com/jojomi/go-script"
	"github.com/jojomi/go-script/print"
	"github.com/jojomi/go-spew/spew"
	"github.com/jojomi/minprofile"
	"github.com/jojomi/strtpl"
	"github.com/jojomi/timecalc"
	"github.com/juju/errors"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	flagFamilybookShowConfig     bool
	flagFamilybookProband        string
	flagFamilybookCompile        bool
	flagFamilybookAnonymize      bool
	flagFamilybookNumCompileRuns int
)

type SubtreeConfig struct {
	Root       generations.Person
	Generation int
}

func getFamilybookCommand() *cobra.Command {
	var familybookCmd = cobra.Command{
		Use:   "familybook",
		Short: "creates genealogytrees using (Lua)LaTeX and the awesome genealogytree package",
		Args:  cobra.MinimumNArgs(1),
		Run:   familybookHandler,
	}
	flags := familybookCmd.PersistentFlags()
	flags.BoolVarP(&flagFamilybookShowConfig, "debug-config", "c", false, "show parsed config")
	flags.StringVarP(&flagFamilybookProband, "proband", "p", "", "overwrite proband")
	flags.BoolVarP(&flagFamilybookAnonymize, "anonymize", "a", false, "anonymize data")
	flags.BoolVarP(&flagFamilybookCompile, "compile", "", true, "generate pdf file using lualatex")
	flags.IntVarP(&flagFamilybookNumCompileRuns, "compile-runs", "n", 2, "number of times to call lualatex")

	return &familybookCmd
}

type RenderConfig struct {
	Config       *FamilybookConfig
	Generations  []Generation
	ExtraPersons *generations.PersonList
	EmptyEntries *generations.PersonList
	Options      map[string]interface{}
	Stats        *generations.Stats
	Lineages     *generations.LineageContainer
	SubtreeRoots []SubtreeConfig
}

func familybookHandler(c *cobra.Command, args []string) {
	// https://godoc.org/runtime/pprof
	if flagRootCPUProfileFile != "" {
		f, err := os.Create(flagRootCPUProfileFile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not create CPU profile")
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal().Err(err).Msg("could not start CPU profile")
		}
		defer pprof.StopCPUProfile()
	}
	p := minprofile.NewStarted()

	if flagRootVerbose {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	log.Debug().Strs("args", args).Msg("executing familybookHandler")
	sc := script.NewContext()

	for _, configFile := range args {
		print.Boldf("Config file: %s...\n", configFile)
		log.Debug().Str("path", sc.AbsPath(configFile)).Msgf("loading config")
		if !sc.FileExists(configFile) {
			log.Fatal().Msgf("file not found: %s\n", sc.AbsPath(configFile))
		}
		print.Successln("File found.")

		var config FamilybookConfig
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		print.Boldln("Reading config data...")
		err = yaml.UnmarshalStrict(data, &config)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		log.Debug().Msg("config data okay")
		print.Successln("Config data OK.")
		p.StepP("config parsed")
		config.SetDefaults()
		// setDefaultOutputPath(&config, configFile)

		if flagFamilybookShowConfig {
			spew.Dump(config)
			os.Exit(0)
		}

		print.Boldln("Preparing template execution...")
		templateData := map[string]interface{}{
			"Now": config.GetNow(),
		}
		log.Trace().Interface("template data", templateData).Msg("dynamically evaluating title and attribution in config")
		config.Title = strtpl.MustEval(config.Title, templateData)
		config.Attribution = strtpl.MustEval(config.Attribution, templateData)
		log.Trace().Interface("config", config).Msg("config with set defaults")

		// load config, use it
		print.Successln("Templates generated.")
		database, err := loadDatabases(config.Databases, flagRootDatabaseBaseDir)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load databases")
		}
		p.StepP("databases loaded")

		probandID := flagFamilybookProband
		if probandID == "" {
			probandID = config.Proband
		}
		person, err := database.GetByID(probandID)
		if err != nil {
			log.Fatal().Err(err).Str("search", probandID).Msg("root person not found")
		}
		log.Debug().Interface("person", person).Msg("root person found")

		// add sources from database (database first!)
		config.Sources = append(database.Sources, config.Sources...)

		lineageContainer := generations.NewLineageContainer()
		log.Trace().Msg("computing parent generations...")
		parentGenerations := getGenerations(config.Options, person, lineageContainer, 1, false)
		log.Trace().Interface("value", parentGenerations).Int("length", len(parentGenerations)).Str("l", spew.Sdump(parentGenerations[0])).Msg("parent generations")
		childGenerations := getGenerations(config.Options, person, lineageContainer, -1, true)

		generationList := make([]Generation, 0, len(parentGenerations)+len(childGenerations))
		for _, p := range parentGenerations {
			generationList = append(generationList, p)
		}
		for _, c := range childGenerations {
			generationList = append(generationList, c)
		}

		for i := range generationList {
			g := &generationList[i]
			if g.Index == 0 {
				g.IsTarget = true
			}
			if g.IsParent && g.Index == len(parentGenerations)-1 {
				g.IsBase = true
			}
			if !g.IsParent && g.Index == len(childGenerations) {
				g.IsBase = true
			}
		}

		if config.Options.IncludePartners {
			for i, g := range generationList {
				generationList[i] = getWithPartners(g, lineageContainer)
			}
		}

		// stats := generations.Stats{}
		// pers := allPersons(generationList)
		// stats.ChildrenCount(pers)
		// os.Exit(1)

		extraPersons := generations.NewPersonList()
		for _, e := range config.ExtraPersons {
			extraPerson, err := database.GetByID(e)
			if err != nil {
				panic(err)
			}
			setSingleTreePositionID(extraPerson)
			extraPersons.AddPerson(extraPerson)
		}

		emptyPersons := generations.NewPersonList()
		for i := 0; i < config.Options.EmptyEntries; i++ {
			emptyPerson := generations.FlatPerson{}
			emptyPersons.AddPerson(&emptyPerson)
		}

		// put all indirectly connected persons to the "extra group"?
		if config.Options.IncludeIslandAsExtra {
			direct := allPersons(generationList)
			allIslandPersons := generations.NewPersonList()
			allIslandPersons.AddPerson(person)
			allNew := generations.NewPersonList()
			allNew.AddList(&allIslandPersons)
			islandSize := allIslandPersons.Count()
			var (
				newParents, newChildren, newPartners generations.PersonList
				reallyNew                            *generations.PersonList
				err                                  error
			)

			stopPersons := []string{
				"MitlMare1988", // Mlynek-Äste
				"MitlAnto1957", // Mitlmeier-Ast (Achtung, das braucht auch StahHube1951!)
				// "StahMari1947", "HofeKlar1949", "StahBrun1950", "StahHube1951", "DeißRena1954", "GödeMoni1956", "BrauRegi1958", "StahGert1960",
				/// Simple cut: Alle Geschwister von Brigitte
				"StahMari1947", "HofeKlar1949", "StahBrun1950", "StahHube1951", "DeißRena1954", "GödeMoni1956", "BrauRegi1958", "StahGert1960", "MitlHelg1962",
			}
			addPartners := true

			for {
				fmt.Println("island size:", allIslandPersons.Count())
				newParents, err = allNew.GetParents()
				if err != nil {
					panic(err)
				}
				newChildren, err = allNew.GetChildren()
				if err != nil {
					panic(err)
				}
				if addPartners {
					newPartners, err = allNew.GetPartners()
					if err != nil {
						panic(err)
					}
				}
				allNew = generations.NewPersonList()
				allNew.AddList(&newParents).AddList(&newChildren)
				if addPartners {
					allNew.AddList(&newPartners)
				}
				for _, s := range stopPersons {
					rp := generations.FlatPerson{}
					rp.SetID(s)
					allNew.Remove(&rp)
				}
				reallyNew, err = allNew.Diff(&allIslandPersons)
				if err != nil {
					panic(err)
				}
				allNew = *reallyNew
				for _, a := range allNew.GetPersons() {
					fmt.Println(a)
				}

				// fmt.Println("children:", newChildren.Count(), "parents:", newParents.Count(), "all new:", allNew.Count(), "really new:", reallyNew.Count())

				// var t string
				// fmt.Scanf("%s\n", &t)

				allIslandPersons.AddList(reallyNew)
				allIslandPersons.RemoveDuplicatesByBestID()
				if islandSize == allIslandPersons.Count() {
					break
				}
				islandSize = allIslandPersons.Count()
			}
			extraPersons.AddList(&allIslandPersons)
			pExtraPersons, err := extraPersons.Diff(direct)
			if err != nil {
				panic(err)
			}
			extraPersons = *pExtraPersons
			fmt.Println("count extra:", extraPersons.Count())
			// os.Exit(1)
		}
		p.StepP("persons put together")

		if config.OutputFilename == "" {
			filename := filepath.Base(configFile)
			filename = strings.TrimSuffix(filename, ".yml")
			config.OutputFilename = filename
		}

		var renderer FamilybookRenderer
		switch config.Renderer {
		case "pdf":
			renderer = NewFamilybookRendererPDF(config)
		case "html":
			renderer = NewFamilybookRendererHTML(config)
		}

		var outputSourceFilename string
		if renderer.DoesCompile() {
			outputSourceFilename = renderer.GetIntermediateFilename()
		} else {
			outputSourceFilename = renderer.GetOutputFilename()
		}
		config.OutputFilename = renderer.GetOutputFilename()
		print.Boldf("Rendering to file: %s\n", outputSourceFilename)
		err = os.MkdirAll(filepath.Dir(config.OutputFilename), 0750)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		renderConfig := RenderConfig{
			Config:       &config,
			Generations:  generationList,
			ExtraPersons: &extraPersons,
			EmptyEntries: &emptyPersons,
			Options:      config.Templates.Document.Options,
			Stats:        generations.NewStats(config.GetNow()),
			Lineages:     lineageContainer,
		}

		// build the Subtree configs (all persons from every nth generation)
		if !config.Options.Subtrees.Hide {
			fc := []SubtreeConfig{}
			for gen := 0; gen < len(renderConfig.Generations); gen += config.Options.Subtrees.GenerationsPerChart {
				for _, p := range renderConfig.Generations[gen].Persons {
					fc = append(fc, SubtreeConfig{
						Root:       p,
						Generation: gen,
					})
				}
			}
			renderConfig.SubtreeRoots = fc
		}

		err = renderer.Render(outputSourceFilename, config.Templates.Document.Filename, renderConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("rendering failed")
		}

		print.Successln("Output source file written.")
		p.StepP("tex file generated")

		if !flagFamilybookCompile {
			log.Debug().Msg("not compiling because of set flag")
			return
		}
		if renderer.DoesCompile() {
			pCompile := minprofile.NewStarted()
			print.Boldln("Compiling document...")
			err = renderer.Compile(c.PersistentFlags(), func(step string) {
				p.StepP(step)
			})
			if err != nil {
				log.Fatal().Err(err).Msg("compilation error")
			}
			print.Successln("Succesfully compiled.")
			pCompile.StepP("TOTAL COMPILE TIME")
		}

		p.StepP("finished")

		if flagRootOpen {
			print.Successf("Opening output file: %s\n", config.OutputFilename)
			openDocument(config.OutputFilename)
		}
	}
}

// templateConcat allows to concat strings with glue.
// The final output is trimmed.
func templateConcat(glue string, inputs ...string) string {
	// remove empty strings, so that glue is applied more usefully
	cleanInputs := []string{}
	for _, input := range inputs {
		if strings.TrimSpace(input) == "" {
			continue
		}
		cleanInputs = append(cleanInputs, input)
	}
	return strings.Join(cleanInputs, glue)
}

func makeCustomFuncMap(customFuncs map[string]interface{}) template.FuncMap {
	funcMap := sprig.TxtFuncMap()

	funcMap["isSpitzahn"] = isSpitzahn
	funcMap["isVollSpitzahn"] = isVollSpitzahn
	funcMap["sub"] = sub
	funcMap["join"] = betterJoin
	funcMap["concat"] = templateConcat
	funcMap["latexify"] = generations.TemplateLatexifyFunc
	funcMap["safeHTML"] = safeHTML
	funcMap["labelify"] = latexLabelify
	funcMap["collapseWhitespace"] = collapseWhitespace
	funcMap["formatDate"] = formatDate
	funcMap["formatDateDayMonth"] = formatDateDayMonth
	funcMap["formatDateYear"] = formatDateYear
	funcMap["formatDateText"] = formatDateText
	funcMap["age"] = templateAge
	funcMap["duration"] = templateDuration
	funcMap["weekday"] = weekday
	funcMap["zeroDate"] = func() time.Time { return time.Time{} }
	funcMap["getFullReferenceLabel"] = getFullReferenceLabel
	funcMap["expandPath"] = expandPath
	funcMap["debugf"] = debugf
	funcMap["fanchartParentTree"] = fanchartParentTree
	funcMap["positionParentTree"] = positionParentTree

	for key, f := range customFuncs {
		funcMap[key] = f
	}
	return funcMap
}

func safeHTML(input string) htmlTemplate.HTML {
	return htmlTemplate.HTML(input)
}

func fanchartParentTree(p generations.Person, numLevels, startLevel, maxLevel int) (string, error) {
	pOpts := &generations.RenderPersonOptions{}
	pOpts.SetDefaults()
	o := generations.RenderTreeOptions{
		MaxParentGenerations:               int(math.Min(float64(numLevels), float64(maxLevel-startLevel))),
		RenderPersonOptions:                pOpts,
		TemplateFilenameTree:               "templates/tree.tpl",
		TemplateFilenamePerson:             "templates/person.tpl",
		TemplateFilenameParentTree:         "templates/parent_tree.tpl",
		TemplateFilenameParentTreeHeadless: "templates/parent_tree_headless.tpl",
		TemplateFilenameChildTree:          "templates/child_tree.tpl",
		TemplateFilenameUnionTree:          "templates/union_tree.tpl",
	}
	o.SetDefaults()
	o.RenderPersonOptions.TemplateFilename = o.TemplateFilenamePerson
	result, err := generations.RenderParentSubtree(p, o, false)
	return string(result), err
}

func debugf(format string, inputs ...interface{}) string {
	log.Debug().Msgf(format, inputs...)
	return ""
}

func expandPath(input string) (string, error) {
	return homedir.Expand(input)
}

func getFullReferenceLabel(sources []generations.Source, searchReference generations.Reference) string {
	for _, source := range sources {
		if source.ID == "" {
			continue
		}
		if source.ID != searchReference.ID {
			continue
		}
		return latexLabelify("src:full", source.ID)
	}
	return ""
}

func latexLabelify(parts ...string) string {
	r := regexp.MustCompile(`\W`)
	glue := "-"
	for i, p := range parts {
		// first element unchanged
		if i == 0 {
			continue
		}
		parts[i] = r.ReplaceAllString(p, "")
	}
	return strings.Join(parts, glue)
}

func isSpitzahn(person generations.Person, allPersons *generations.PersonList) bool {
	m, err := person.GetMom()
	if err != nil {
		return true
	}
	if m == nil || m.IsDummy() || !allPersons.Contains(m) {
		return true
	}
	d, err := person.GetDad()
	if err != nil {
		return true
	}
	if d == nil || d.IsDummy() || !allPersons.Contains(d) {
		return true
	}

	return false
}

func isVollSpitzahn(person generations.Person, allPersons *generations.PersonList) bool {
	m, err := person.GetMom()
	if err != nil {
		return true
	}
	if m != nil && !m.IsDummy() && allPersons.Contains(m) {
		return false
	}
	d, err := person.GetDad()
	if err != nil {
		return true
	}
	if d != nil && !d.IsDummy() && allPersons.Contains(d) {
		return false
	}

	return true
}

func sparseDate(input string) *timecalc.SparseDate {
	return timecalc.SparseDateFrom(input)
}

func getWithPartners(g Generation, lineageContainer *generations.LineageContainer) Generation {
	pList := generations.PersonList{}

	var lastInLineage generations.Person
	for _, p := range g.Persons {
		lineage := lineageContainer.Get(p.GetMainTreePositionID())
		if lineage != nil {
			lastInLineage = lineage.GetPersons()[lineage.Count()-1]
		} else {
			lastInLineage = nil
		}
		pList.AddPerson(p)
		partners, err := p.GetPartners()
		if err != nil {
			panic(err)
		}
		for _, partner := range partners.GetPersons() {
			partner = partner.Copy()
			setTreePos(partner, p)

			// Partner already contained *at this tree position* because he is a ancestor too?
			searchPartnerTreeID := getTreePos(partner, lastInLineage)
			if lineageContainer.Get(searchPartnerTreeID) != nil {
				continue
			}
			lineageContainer.Add(partner.GetBestID(), partner.GetMainTreePositionID(), generations.NewPersonList())
			pList.AddPerson(partner)
		}
	}

	result := Generation{
		Index:    g.Index,
		IsParent: g.IsParent,
		Name:     g.Name,
		Persons:  pList.GetPersons(),
	}
	return result
}

func percentageInt(total, part, significantPlaces int) string {
	return percentage(float64(total), float64(part), significantPlaces)
}

func percentage(total, part float64, significantPlaces int) string {
	p := total / part * 100.0
	return fmt.Sprintf("%."+strconv.Itoa(significantPlaces)+"f", p)
}

func renderFamilybookEntry(config FamilybookConfig, person generations.Person, lineages *generations.LineageContainer, allPersons *generations.PersonList) (string, error) {
	templateFilename := config.Templates.Entry.Filename
	templateContent, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		return "", err
	}

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
		Config:              &config,
		Person:              person,
		Lineages:            lineages,
		ZeroDate:            time.Time{},
		AllPersons:          allPersons,
		ReferencesContainer: referencesContainer,
	}
	funcMap := makeCustomFuncMap(map[string]interface{}{
		"findTreePositionID": findTreePositionID,
		"derefInt":           derefInt,
		"derefUint":          derefUint,
		"uint2int":           uint2int,
		"refName":            refName,
		"placeConnection":    placeConnection,
		"printDuration":      printDuration,
	})
	result, err := strtpl.EvalWithFuncMap(string(templateContent), funcMap, data)
	if err != nil {
		return "", errors.Annotatef(err, "Rendering person: %v", person)
	}
	return result, nil
}

func findTreePositionID(new, base generations.Person, lineages *generations.LineageContainer) string {
	baseID := ""
	if base != nil {
		baseID = base.GetMainTreePositionID()
	}
	newID := baseID + "-" + new.GetBestID()
	if match := lineages.Get(newID); match == nil {
		baseID = regexp.MustCompile(`-[^-]+$`).ReplaceAllString(baseID, "")
		newID = baseID + "-" + new.GetBestID()
	}
	if match := lineages.Get(newID); match == nil {
		baseID = regexp.MustCompile(`-[^-]+$`).ReplaceAllString(baseID, "")
		newID = baseID + "-" + new.GetBestID()
	}
	if match := lineages.Get(newID); match != nil {
		new.ClearTreePositionIDs().AddTreePositionID(newID)
	}
	return ""
}

func getTreePos(new, base generations.Person) string {
	baseID := ""
	if base != nil {
		baseID = base.GetMainTreePositionID()
	}
	return baseID + "-" + new.GetBestID()
}

func setTreePos(new, base generations.Person) string {
	new.ClearTreePositionIDs().AddTreePositionID(getTreePos(new, base))
	return ""
}

func betterJoin(sep string, data ...string) string {
	return strings.Join(data, sep)
}

func allPersons(g []Generation, lists ...*generations.PersonList) *generations.PersonList {
	result := generations.PersonList{}

	for _, generation := range g {
		result.AddSlice(filterDummies(generation.Persons))
	}

	for _, l := range lists {
		result.AddList(l)
	}

	return &result
}

func sortByBirthOrBaptism(persons *generations.PersonList) *generations.PersonList {
	return persons.GetFiltered(
		func(p generations.Person) bool {
			// valid?
			if p.GetBirth().Date != "" {
				return true
			}
			if p.GetBaptism().Date != "" {
				return true
			}
			return false
		},
	).
		RemoveDuplicatesByBestID().
		SortPersons(
			func(a, b generations.Person) bool {
				var dateA, dateB *timecalc.SparseDate
				if a.GetBirth().Date != "" {
					dateA = timecalc.SparseDateFrom(a.GetBirth().Date)
				}
				if dateA == nil && a.GetBaptism().Date != "" {
					dateA = timecalc.SparseDateFrom(a.GetBaptism().Date)
				}

				if b.GetBirth().Date != "" {
					dateB = timecalc.SparseDateFrom(b.GetBirth().Date)
				}
				if dateB == nil && b.GetBaptism().Date != "" {
					dateB = timecalc.SparseDateFrom(b.GetBaptism().Date)
				}

				return dateA.GetMinDate().Before(dateB.GetMinDate())
			},
		)
}

func sortByDeath(persons *generations.PersonList) *generations.PersonList {
	return persons.GetFiltered(
		func(p generations.Person) bool {
			// valid?
			return p.GetDeath().Date != ""
		},
	).
		RemoveDuplicatesByBestID().
		SortPersons(
			func(a, b generations.Person) bool {
				var dateA, dateB *timecalc.SparseDate
				if a.GetDeath().Date != "" {
					dateA = timecalc.SparseDateFrom(a.GetDeath().Date)
				}

				if b.GetDeath().Date != "" {
					dateB = timecalc.SparseDateFrom(b.GetDeath().Date)
				}

				return dateA.GetMinDate().Before(dateB.GetMinDate())
			},
		)
}

func titleName(name generations.Name) string {
	return collapseWhitespace(strtpl.MustEval(`{{ .GetUsedFirst }} {{ .Last }}`, name))
}

func subData(params ...interface{}) interface{} {
	var data interface{}
	if len(params) == 1 {
		data = params[0]
	} else {
		dataMap := make(map[string]interface{})
		for i := 0; i < len(params)-1; i = i + 2 {
			dataMap[params[i].(string)] = params[i+1]
		}
		data = dataMap
	}
	return data
}

func subHTML(name string, params ...interface{}) (htmlTemplate.HTML, error) {
	data := subData(params...)

	templateFilename := filepath.Join("templates", "html", "sub", name)
	templateContent, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		return "", err
	}
	customFuncMap := makeCustomFuncMap(map[string]interface{}{
		// type assertion funcs
		"asPersonListPointer": asPersonListPointer,
		"zeroDate":            func() time.Time { return time.Time{} },
		"derefInt":            derefInt,
		"derefUint":           derefUint,
		"uint2int":            uint2int,
		"refName":             refName,
		"placeConnection":     placeConnection,
		"printDuration":       printDurationHTML,
	})
	customFuncMap["sub"] = subHTML
	funcMap := makeHTMLFuncMap(symbolsHTML(), spacingHelpersHTML(), toHTMLFuncMap(customFuncMap), personHelpers())
	result, err := strtpl.EvalHTMLWithFuncMap(string(templateContent), funcMap, data)
	if err != nil {
		return "", errors.Annotatef(err, "Rendering subtemplate: %v", name)
	}
	return htmlTemplate.HTML(result), nil
}

func sub(name string, params ...interface{}) (string, error) {
	data := subData(params...)

	templateFilename := filepath.Join("templates", "sub", name)
	templateContent, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		return "", err
	}
	funcMap := makeCustomFuncMap(map[string]interface{}{
		// type assertion funcs
		"asPersonListPointer": asPersonListPointer,
		"zeroDate":            func() time.Time { return time.Time{} },
		"derefInt":            derefInt,
		"derefUint":           derefUint,
		"uint2int":            uint2int,
		"refName":             refName,
		"placeConnection":     placeConnection,
		"printDuration":       printDuration,
	})
	result, err := strtpl.EvalWithFuncMap(string(templateContent), funcMap, data)
	if err != nil {
		return "", errors.Annotatef(err, "Rendering subtemplate: %v", name)
	}
	return result, nil
}

func weekday(input string) string {
	if !isSparseDateFormat(input) {
		return ""
	}

	s := timecalc.SparseDateFrom(input)
	if !s.IsComplete() {
		return ""
	}
	d := s.GetDate()

	switch d.Weekday() {
	case time.Monday:
		return "Montag"
	case time.Tuesday:
		return "Dienstag"
	case time.Wednesday:
		return "Mittwoch"
	case time.Thursday:
		return "Donnerstag"
	case time.Friday:
		return "Freitag"
	case time.Saturday:
		return "Samstag"
	case time.Sunday:
		return "Sonntag"
	}
	return ""
}

func placeConnection(person generations.Person, places ...string) bool {
	for _, place := range places {
		datePlaces := []generations.DatePlace{
			person.GetBirth(),
			person.GetBaptism(),
			person.GetDeath(),
			person.GetBurial(),
		}
		for _, datePlace := range datePlaces {
			if strings.Contains(datePlace.Place, place) {
				return true
			}
		}

		// residences
		for _, residence := range person.GetResidences() {
			if residence.City == place {
				return true
			}
		}
	}
	return false
}

type DurationTexts struct {
	Year   string
	Years  string
	Month  string
	Months string
	Day    string
	Days   string
}

func printDuration(duration timecalc.Duration, options FamilybookOptions, grammarCase string) (string, error) {
	/// return "printdur", nil
	textMap := map[string]DurationTexts{
		"nominativ": {
			Year:   "Jahr",
			Years:  "Jahre",
			Month:  "Monat",
			Months: "Monate",
			Day:    "Tag",
			Days:   "Tage",
		},
		"dativ": {
			Year:   "Jahr",
			Years:  "Jahren",
			Month:  "Monat",
			Months: "Monaten",
			Day:    "Tag",
			Days:   "Tagen",
		},
	}
	texts, ok := textMap[grammarCase]
	if !ok {
		return "", errors.New("Grammar case invalid: " + grammarCase)
	}

	data := struct {
		Duration timecalc.Duration
		Texts    DurationTexts
	}{
		Duration: duration,
		Texts:    texts,
	}

	var (
		err      error
		yearStr  string
		monthStr string
		dayStr   string
	)
	switch {
	case duration.Years == 1:
		yearStr, err = strtpl.Eval(`{{- .Duration.Years }}~{{ .Texts.Year -}}`, data)
	case duration.Years > 1:
		yearStr, err = strtpl.Eval(`{{- .Duration.Years }}~{{ .Texts.Years -}}`, data)
	}

	switch {
	case duration.Months == 1:
		monthStr, err = strtpl.Eval(`{{- .Duration.Months }}~{{ .Texts.Month -}}`, data)
	case duration.Months > 1:
		monthStr, err = strtpl.Eval(`{{- .Duration.Months }}~{{ .Texts.Months -}}`, data)
	}

	switch {
	case duration.Days == 1:
		dayStr, err = strtpl.Eval(`{{- .Duration.Days }}~{{ .Texts.Day -}}`, data)
	case duration.Days > 1:
		dayStr, err = strtpl.Eval(`{{- .Duration.Days }}~{{ .Texts.Days -}}`, data)
	}

	if err != nil {
		return "", errors.Annotate(err, "Duration template failed")
	}

	parts := []string{}
	if yearStr != "" {
		parts = append(parts, yearStr)
	}
	if monthStr != "" {
		parts = append(parts, monthStr)
	}
	if dayStr != "" {
		parts = append(parts, dayStr)
	}

	result := strings.Join(parts, ", ")

	if !duration.IsFinished {
		result += "*"
	}

	return result, nil
}

func printDurationHTML(duration timecalc.Duration, options FamilybookOptions, grammarCase string) (string, error) {
	data, err := printDuration(duration, options, grammarCase)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(data, "~", " "), nil
}

func refName(input string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9]`)
	return re.ReplaceAllString(input, `-`)
}

func uint2int(i uint) int {
	return int(i)
}

func derefInt(i *int) int {
	return *i
}

func derefUint(i *uint) uint {
	return *i
}

func asPersonListPointer(input interface{}) *generations.PersonList {
	return input.(*generations.PersonList)
}

// func nameFullNoBirth(name generations.Name) string {
// 	return collapseWhitespace(strtpl.MustEval(
// 		`{{ $used := .GetUsedFirst }}
// 		{{ range .First }}
// 			{{ if eq . $used }}
// 				\usedFirstname{ {{- . -}} }
// 			{{ else }}
// 				\additionalFirstname{ {{- . -}} }
// 			{{ end }}
// 		{{ end }}
// 		{{- with .Nick -}}
// 			{{ if . }}
// 			\nickname{ {{- . -}} }
// 			{{- end }}
// 		{{- end }}
// 		{{ if .Last -}}
// 		\lastname{ {{- .Last -}} }
// 		{{- end }}
// 		`, name))
// }

// func nameFullReversed(name generations.Name) string {
// 	return collapseWhitespace(strtpl.MustEval(
// 		`{{ if .Last -}}
// 		\lastname{ {{- .Last -}} }
// 		{{- else -}}
// 		?
// 		{{- end }},
// 		{{ $used := .GetUsedFirst }}
// 		{{ range .First }}
// 			{{ if eq . $used }}
// 				\usedFirstname{ {{- . -}} }
// 			{{ else }}
// 				\additionalFirstname{ {{- . -}} }
// 			{{ end }}
// 		{{ end }}
// 		{{- with .Nick -}}
// 			{{ if . }}
// 			\nickname{ {{- . -}} }
// 			{{- end }}
// 		{{- end }}
// 		`, name))
// }

// func nameBirth(name generations.Name) string {
// 	return collapseWhitespace(strtpl.MustEval(
// 		`\firstname{ {{- .GetUsedFirst -}} }
// 		{{- with .Nick -}}
// 			{{ if . }}
// 			\nickname{ {{- . -}} }
// 			{{- end }}
// 		{{- end }}
// 		\lastname{ {{- .Last -}} }
// 		{{- with .Birth -}}
// 			{{ if . -}}
// 			\birthname{ {{- . -}} }
// 			{{- end }}
// 		{{- end -}}
// 		{{ if .Alias }}
// 		\aliasname{ {{- .Alias -}} }
// 		{{- end -}}
// 		`, name))
// }

func collapseWhitespace(input string) string {
	s := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(s.ReplaceAllString(input, " "))
}

func templateDuration(from, to string, now time.Time) timecalc.Duration {
	result := timecalc.Duration{}
	if !isSparseDateFormat(from) {
		result.Error = errors.Errorf("could not parse from date")
		return result
	}
	fromDate := timecalc.SparseDateFrom(from)
	if !fromDate.IsComplete() {
		result.Error = errors.Errorf("from date is not complete")
		return result
	}
	var toFull time.Time
	if to != "" {
		if !isSparseDateFormat(to) {
			result.Error = errors.Errorf("could not parse to date")
			return result
		}
		toDate := timecalc.SparseDateFrom(to)
		if !toDate.IsComplete() {
			result.Error = errors.Errorf("from date is not complete")
			return result
		}
		toFull = toDate.GetDate()
		result.IsFinished = true
	} else {
		if now.IsZero() {
			result.Error = errors.Errorf("no valid end, but also no now given")
			return result
		}
		toFull = now
	}
	fromFull := fromDate.GetDate()
	a := timecalc.FullDate{&fromFull}.Age(toFull)
	result.Years = a.Years
	result.Months = a.Months
	result.Days = a.Days
	return result
}

func templateAge(then, now string) int {
	errorResult := -1
	if !isSparseDateFormat(then) {
		return errorResult
	}
	thenDate := timecalc.SparseDateFrom(then)
	if !thenDate.IsComplete() {
		return errorResult
	}
	if !isSparseDateFormat(now) {
		return errorResult
	}
	nowDate := timecalc.SparseDateFrom(now)
	if !nowDate.IsComplete() {
		return errorResult
	}
	return age.AgeAt(thenDate.GetDate(), nowDate.GetDate())
}

func formatDate(input string) string {
	if !isSparseDateFormat(input) {
		return input
	}

	s := timecalc.SparseDateFrom(input)
	return s.Format("02.01.2006")
}

func formatDateYear(input string) string {
	if !isSparseDateFormat(input) {
		return input
	}

	s := timecalc.SparseDateFrom(input)
	return s.Format("2006")
}

func formatDateDayMonth(input string) string {
	if !isSparseDateFormat(input) {
		return input
	}

	s := timecalc.SparseDateFrom(input)
	return s.Format("02.01.")
}

func formatDateText(input string) string {
	if !isSparseDateFormat(input) {
		return input
	}

	s := timecalc.SparseDateFrom(input)
	month := "?"
	switch s.Format("1") {
	case "1":
		month = "Jan."
	case "2":
		month = "Feb."
	case "3":
		month = "März"
	case "4":
		month = "Apr."
	case "5":
		month = "Mai"
	case "6":
		month = "Juni"
	case "7":
		month = "Juli"
	case "8":
		month = "Aug."
	case "9":
		month = "Sept."
	case "10":
		month = "Okt."
	case "11":
		month = "Nov."
	case "12":
		month = "Dez."
	}

	return s.Format("2. ") + month + s.Format(" 2006")
}

func isSparseDateFormat(input string) bool {
	re := regexp.MustCompile(`[\d?]{4}-[\d?]{2}-[\d?]{2}`)
	return re.MatchString(input)
}

func getGenerations(config FamilybookOptions, p generations.Person, lineageContainer *generations.LineageContainer, direction int, hideRoot bool) []Generation {
	pList := generations.PersonList{}
	pList.AddPerson(p)
	setSingleTreePositionID(p)

	if config.IncludeProbandSiblings {
		log.Trace().Msg("adding proband siblings")
		siblings, err := p.GetSiblings()
		if err != nil {
			log.Fatal().Err(err).Interface("proband", p).Msg("could not retrieve proband's siblings")
		}
		if siblings.Common != nil {
			// attention: only full consindered siblings here!
			for _, sibling := range siblings.Common.GetPersons() {
				log.Trace().Interface("person", sibling).Msg("adding proband sibling")
				sibling.SetStopPerson(true)
				pList.AddPerson(sibling)
				setSingleTreePositionID(sibling)
			}
		}
	}
	return getGenerationsRecursive(0, config, pList, lineageContainer, direction, hideRoot)
}

func setSingleTreePositionID(p generations.Person) {
	//newList := generations.NewPersonList()
	//treePositionID := newList.AddPerson(p).GetTreePositionID()
	treePositionID := "-" + p.GetBestID()
	p.AddTreePositionID(treePositionID)
}

func filterDummies(persons []generations.Person) []generations.Person {
	result := make([]generations.Person, 0, len(persons))
	for _, p := range persons {
		if p.IsDummy() {
			continue
		}
		result = append(result, p)
	}
	return result
}

func getGenerationsRecursive(level int, config FamilybookOptions, persons generations.PersonList, lineageContainer *generations.LineageContainer, direction int, hideRoot bool) []Generation {
	result := make([]Generation, 0)
	if !hideRoot {
		genName := GetGenerationName(level * direction)
		if level == 0 && config.Texts.Proband != "" {
			genName = config.Texts.Proband
		}
		pg := Generation{
			Index:   level,
			Name:    genName,
			Persons: persons.GetPersons(),
		}
		pg.IsParent = direction == 1
		result = append(result, pg)
	}
	if direction > 0 && level >= config.MaxParentGenerations {
		log.Trace().Int("value", config.MaxParentGenerations).Msg("cutoff because MaxParentGenerations was reached (see config)")
		return result
	}
	if direction < 0 && level >= config.MaxChildGenerations {
		log.Trace().Int("value", config.MaxChildGenerations).Msg("cutoff because MaxChildGenerations was reached (see config)")
		return result
	}
	// parents?
	var (
		nextLevel generations.PersonList
		err       error
	)
	if direction > 0 {
		for _, person := range persons.GetPersons() {
			if person.IsDummy() {
				continue
			}
			if person.IsStopPerson() {
				continue
			}
			mom, err := person.GetMom()
			if err != nil {
				panic(err)
			}
			log.Trace().Str("person", person.String()).Str("mom", mom.String()).Msg("found mom")
			dad, err := person.GetDad()
			if err != nil {
				panic(err)
			}
			log.Trace().Str("person", person.String()).Str("dad", dad.String()).Msg("found dad")
			mom = mom.Copy()
			dad = dad.Copy()

			pID := person.GetMainTreePositionID()

			if !mom.IsDummy() {
				setTreePos(mom, person)
			}
			if !dad.IsDummy() {
				setTreePos(dad, person)
			}

			if config.MaternalOrder {
				if !mom.IsDummy() {
					lineageContainer.AddWithBase(mom.GetBestID(), mom.GetMainTreePositionID(), pID, person)
					nextLevel.AddPerson(mom)
				}
			}
			if !dad.IsDummy() {
				lineageContainer.AddWithBase(dad.GetBestID(), dad.GetMainTreePositionID(), pID, person)
				nextLevel.AddPerson(dad)
			}
			if !config.MaternalOrder {
				if !mom.IsDummy() {
					lineageContainer.AddWithBase(mom.GetBestID(), mom.GetMainTreePositionID(), pID, person)
					nextLevel.AddPerson(mom)
				}
			}
		}
	} else {
		for _, person := range persons.GetPersons() {
			if person.IsDummy() {
				continue
			}
			pID := person.GetMainTreePositionID()
			children, err := person.GetChildren()
			if err != nil {
				panic(err)
			}
			/*var (
				childTreePositionID string
				lineage             = lineageContainer.Get(person.GetBestID())
			)*/
			for _, child := range children.GetPersons() {
				if child.IsDummy() {
					continue
				}

				child = child.Copy()

				/*if lineage != nil {
					childTreePositionID = lineage.GetTreePositionIDWith(child)
				} else {
					newList := generations.NewPersonList()
					childTreePositionID = newList.AddPerson(child).GetTreePositionID()
				}
				child.AddTreePositionID(childTreePositionID)*/
				setTreePos(child, person)
				lineageContainer.AddWithBase(child.GetBestID(), child.GetMainTreePositionID(), pID, person)
				nextLevel.AddPerson(child)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	if len(nextLevel.GetPersons()) == 0 {
		return result
	}
	/// TODO not like this: nextLevel.RemoveDuplicates()
	childResult := getGenerationsRecursive(level+1, config, nextLevel, lineageContainer, direction, false)
	for _, r := range childResult {
		result = append(result, r)
	}
	return result
}

func GetGenerationName(level int) string {
	// source: http://www.der-familienstammbaum.de/ahnenforschung/familienstammbaum-erstellen
	switch level {
	case -6:
		return "Ururururenkel"
	case -5:
		return "Urururenkel"
	case -4:
		return "Ururenkel"
	case -3:
		return "Urenkel"
	case -2:
		return "Enkel"
	case -1:
		return "Kinder"
	case 0:
		return "Proband"
	case 1:
		return "Eltern"
	case 2:
		return "Großeltern"
	case 3:
		return "Urgroßeltern"
	case 4:
		return "Alteltern"
	case 5:
		return "Altgroßeltern"
	case 6:
		return "Alturgroßeltern"
	case 7:
		return "Obereltern"
	case 8:
		return "Obergroßeltern"
	case 9:
		return "Oberurgroßeltern"
	case 10:
		return "Stammeltern"
	case 11:
		return "Stammgroßeltern"
	case 12:
		return "Stammurgroßeltern"
	case 13:
		return "Ahneneltern"
	case 14:
		return "Ahnengroßeltern"
	case 15:
		return "Ahnenurgroßeltern"
	case 16:
		return "Urahneneltern"
	case 17:
		return "Urahnengroßeltern"
	case 18:
		return "Urahnenurgroßeltern"
	case 19:
		return "Erzeltern"
	case 20:
		return "Erzgroßeltern"
	case 21:
		return "Erzurgroßeltern"
	case 22:
		return "Erzahneneltern"
	case 23:
		return "Erzahnengroßeltern"
	case 24:
		return "Erzahnenurgroßeltern"
	}

	// TODO fallback using ur
	if level > 0 {
		return fmt.Sprintf("%d. Elterngeneration", level)
	} else {
		return fmt.Sprintf("%d. Kindgeneration", -level)
	}
}

type Generation struct {
	IsParent  bool
	IsBase    bool
	IsTarget  bool
	Index     int
	Name      string
	Persons   []generations.Person
}
