package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/jojomi/generations"
	"github.com/jojomi/go-script/print"
	script "github.com/jojomi/go-script/v2"
	"github.com/jojomi/go-spew/spew"
	"github.com/jojomi/strtpl"
	"github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	flagGenealogytreeShowConfig     bool
	flagGenealogytreeCompile        bool
	flagGenealogytreeAnonymize      bool
	flagGenealogytreeNumCompileRuns int
	flagGenealogytreeMinify         bool
	flagGenealogytreeCheckIDs       bool
)

func getGenealogytreeCommand() *cobra.Command {
	var genealogytreeCmd = cobra.Command{
		Use:   "genealogytree",
		Short: "creates genealogytrees using (Lua)LaTeX and the awesome genealogytree package",
		Args:  cobra.MinimumNArgs(1),
		Run:   genealogytreeHandler,
	}
	flags := genealogytreeCmd.PersistentFlags()
	flags.BoolVarP(&flagGenealogytreeShowConfig, "debug-config", "c", false, "show parsed config")
	flags.BoolVarP(&flagGenealogytreeCheckIDs, "check-ids", "i", true, "error on unlinked IDs")
	flags.BoolVarP(&flagGenealogytreeAnonymize, "anonymize", "a", false, "anonymize data")
	flags.BoolVarP(&flagGenealogytreeCompile, "compile", "", true, "generate pdf file using lualatex")
	flags.IntVarP(&flagGenealogytreeNumCompileRuns, "compile-runs", "n", 2, "number of times to call lualatex")
	flags.BoolVarP(&flagGenealogytreeMinify, "minify", "m", true, "minify filesize of generated pdf file")

	genealogytreeCmd.AddCommand(getTestCommand())

	return &genealogytreeCmd
}

func genealogytreeHandler(c *cobra.Command, args []string) {
	sc := script.NewContext()

	for _, configFile := range args {
		print.Boldf("Config file: %s...\n", configFile)
		if !sc.FileExists(configFile) {
			log.Fatal().Str("config filename", sc.AbsPath(configFile)).Msg("file not found")
		}
		print.Successln("File found.")

		var config GenealogytreeConfig
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		print.Boldln("Reding config data...")
		err = yaml.UnmarshalStrict(data, &config)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		print.Successln("Config data OK.")
		config.SetDefaults()
		setDefaultOutputPath(&config, configFile)

		if flagGenealogytreeShowConfig {
			spew.Dump(config)
			os.Exit(1)
		}

		print.Boldln("Preparing template execution...")
		templateData := map[string]interface{}{
			"Now":  time.Now(),
			"Date": config.Date,
		}
		config.Title = strtpl.MustEval(config.Title, templateData)
		config.Attribution = strtpl.MustEval(config.Attribution, templateData)

		// load config, use it
		for i, treeConfig := range config.Trees {
			treeConfig.AddGlobals(config)

			// level handling
			treeConfig.Levels.AddDefaultLevels(-20, 20)
			treeConfig.Levels.Inherit(treeConfig.ProbandLevel, config.Levels)
			// reverse order!
			themes := treeConfig.Levels.Themes
			for i := range themes {
				theme := themes[len(themes)-1-i]
				themePath := filepath.Join("templates", "levels", theme+".yml")
				themeData, err := ioutil.ReadFile(themePath)
				if err != nil {
					log.Fatal().Err(err).Msg("")
				}
				var themeLevels generations.LevelConfig
				err = yaml.Unmarshal(themeData, &themeLevels)
				if err != nil {
					log.Fatal().Err(err).Msg("")
				}
				themeLevels.AddDefaultLevels(-20, 20)
				themeLevels.Combine(treeConfig.ProbandLevel)
				treeConfig.Levels.Inherit(treeConfig.ProbandLevel, themeLevels)
			}
			treeConfig.Levels.Combine(treeConfig.ProbandLevel)

			database, err := loadDatabases(treeConfig.Databases, flagRootDatabaseBaseDir)

			if err != nil {
				log.Fatal().Err(err).Msg("failed to load databases")
			}

			if flagGenealogytreeAnonymize {
				database.Anonymize()
			}

			o := treeConfig.RenderTreeOptions
			o.Levels = treeConfig.Levels.Combined
			// template filenames
			if o.TemplateFilenameTree == "" {
				o.TemplateFilenameTree = "templates/tree.tpl"
			}
			if o.TemplateFilenamePerson == "" {
				o.TemplateFilenamePerson = "templates/person.tpl"
			}
			if o.TemplateFilenameParentTree == "" {
				o.TemplateFilenameParentTree = "templates/parent_tree.tpl"
			}
			if o.TemplateFilenameParentTreeHeadless == "" {
				o.TemplateFilenameParentTreeHeadless = "templates/parent_tree_headless.tpl"
			}
			if o.TemplateFilenameChildTree == "" {
				o.TemplateFilenameChildTree = "templates/child_tree.tpl"
			}
			if o.TemplateFilenameUnionTree == "" {
				o.TemplateFilenameUnionTree = "templates/union_tree.tpl"
			}
			if o.RenderPersonOptions == nil {
				o.RenderPersonOptions = &generations.RenderPersonOptions{}
			}
			o.RenderPersonOptions.TemplateFilename = o.TemplateFilenamePerson
			o.RenderPersonOptions.Date = treeConfig.Date

			person, err := database.Get(treeConfig.Proband)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			tree, err := generations.RenderGenealogytree(person, o)
			if err != nil {
				fmt.Println(err)
				os.Exit(3)
			}

			treeConfig.Content = string(tree)

			config.Trees[i] = treeConfig
		}

		var renderedTrees string
		for _, treeConfig := range config.Trees {
			renderedTree, err := generations.RenderTemplateFile(treeConfig.Templates.Tree.Filename, struct {
				Config     GenealogytreeConfig
				TreeConfig TreeConfig
				Options    map[string]interface{}
			}{
				Config:     config,
				TreeConfig: treeConfig,
				Options:    treeConfig.Templates.Tree.Options,
			}, template.FuncMap{})
			if err != nil {
				fmt.Println(err)
				os.Exit(6)
			}
			renderedTrees = renderedTrees + string(renderedTree)
		}
		config.RenderedTrees = string(renderedTrees)
		print.Successln("Templates generated.")

		print.Boldf("Rendering to file: %s\n", config.OutputFilename)
		err = os.MkdirAll(filepath.Dir(config.OutputFilename), 0750)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
		renderConfig := struct {
			Config  GenealogytreeConfig
			Options map[string]interface{}
		}{
			config,
			config.Templates.Document.Options,
		}
		err = renderDocument(config.Templates.Document.Filename, renderConfig, template.FuncMap{}, strings.Replace(config.OutputFilename, ".pdf", ".tex", -1))
		if err != nil {
			fmt.Println(err)
			os.Exit(5)
		}
		print.Successln("Output source file written.")

		if !flagGenealogytreeCompile {
			os.Exit(0)
		}
		print.Boldln("Compiling document (lualatex)...")
		err = compileDocument(
			strings.Replace(config.OutputFilename, ".pdf", ".tex", -1),
			flagGenealogytreeNumCompileRuns,
			flagRootVerbose,
			func(step string) {},
		)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		print.Successln("Sucesfully compiled.")

		var openFilename = config.OutputFilename
		if flagGenealogytreeMinify {
			print.Boldln("Minifying output PDF...")
			smallFilename := strings.Replace(config.OutputFilename, ".pdf", ".small.pdf", -1)
			err = minifyDocument(
				config.OutputFilename,
				smallFilename,
			)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}
			print.Successln("Sucesfully minified.")
			openFilename = smallFilename
		}

		if flagRootOpen {
			print.Successf("Opening output file: %s\n", openFilename)
			openDocument(openFilename)
		}
	}
}

func compileDocument(inputFile string, numRuns int, verbose bool, stepFunc func(step string)) error {
	sc := script.NewContext()
	if !sc.CommandExists("lualatex") {
		return fmt.Errorf("lualatex not found in PATH")
	}
	dir := filepath.Dir(inputFile)

	lualatexEarlyCommand := script.LocalCommandFrom("lualatex --interaction=nonstopmode --shell-escape")
	lualatexEarlyCommand.AddAll(
		//"--aux-directory="+dir,
		"--output-directory="+dir,
		"-halt-on-error",
	)
	// clone
	lualatexLastCommand := script.LocalCommandFrom(lualatexEarlyCommand.String())
	lualatexEarlyCommand.AddAll(
		"--draftmode",
		inputFile,
	)
	lualatexLastCommand.AddAll(
		inputFile,
	)

	localCommandMakeindex := script.LocalCommandFrom("makeindex")
	localCommandMakeindex.AddAll(
		strings.Replace(inputFile, ".tex", ".idx", 1),
	)
	execFunc := sc.ExecuteDebug
	if !verbose {
		execFunc = sc.ExecuteSilent
	}

	stepFunc("compilation prep work done")
	var lualatexCommand script.Command
	for i := 0; i < numRuns; i++ {
		lualatexCommand = lualatexEarlyCommand
		// last?
		if i == numRuns-1 {
			lualatexCommand = lualatexLastCommand
		}
		pr, err := execFunc(lualatexCommand)
		if err != nil {
			return err
		}
		if !pr.Successful() {
			return fmt.Errorf("lualatex compilation not successful")
		}
		stepFunc(fmt.Sprintf("lualatex compilation round %d done", i+1))
		pr, err = execFunc(localCommandMakeindex)
		if err != nil {
			return err
		}
		// depending on the number of previous calls to lualatex this might be okay or at least not a reason to throw up
		if !pr.Successful() {
			// return fmt.Errorf("makeindex compilation not successful")
		}
		stepFunc(fmt.Sprintf("makeindex round %d done", i+1))
	}

	return nil
}

func minifyDocument(inputFile, outputFile string) error {
	sc := script.NewContext()
	if !sc.CommandExists("gs") {
		return fmt.Errorf("gs not found in PATH")
	}

	command := strtpl.MustEval(
		`gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.4 -dPDFSETTINGS=/printer -dNOPAUSE -dQUIET -dBATCH -sOutputFile="{{ .OutputFile }}" "{{ .InputFile }}"`,
		struct {
			InputFile  string
			OutputFile string
		}{
			InputFile:  inputFile,
			OutputFile: outputFile,
		},
	)
	localCommand := script.LocalCommandFrom(command)

	pr, err := sc.ExecuteDebug(localCommand)
	if err != nil {
		return err
	}
	if !pr.Successful() {
		return fmt.Errorf("lualatex compilation not successful")
	}

	return nil
}

func openDocument(filename string) {
	open.Start(filename)
}

func setDefaultOutputPath(config OutputFilenameConfig, filename string) {
	if config.GetOutputFilename() != "" {
		return
	}

	f := filepath.Base(filename)
	f = strings.TrimSuffix(f, ".yml")
	config.SetOutputFilename(filepath.Join("output", f+".pdf"))
}

func getTestCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "test",
		Short: "render test document",
		Run:   commandTest,
	}
	return &cmd
}

func commandTest(c *cobra.Command, args []string) {
	var config GenealogytreeConfig

	files, err := ioutil.ReadDir(filepath.Join("..", "testdata", "database"))
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	for _, file := range files {
		if file.Name() == "invalid.yml" {
			continue
		}
		fmt.Printf("building %s...\n", file.Name())

		database := generations.NewMemoryDatabase()

		err := database.ParseYamlFile(filepath.Join("..", "testdata", "database", file.Name()))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		renderOptions := generations.RenderTreeOptions{
			TemplateFilenameTree:               "templates/tree.tpl",
			TemplateFilenamePerson:             "templates/person.tpl",
			TemplateFilenameParentTree:         "templates/parent_tree.tpl",
			TemplateFilenameParentTreeHeadless: "templates/parent_tree_headless.tpl",
			TemplateFilenameChildTree:          "templates/child_tree.tpl",
			TemplateFilenameUnionTree:          "templates/union_tree.tpl",
			GraphType:                          generations.GraphTypeSandclock,
		}

		rootID := "gauss"
		person, err := database.GetByID(rootID)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		tree, err := generations.RenderGenealogytree(person, renderOptions)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		TreeConfig := TreeConfig{
			Content: string(tree),
			Title:   file.Name(),

			PostContent: "{\\tiny \\begin{verbatim}" + string(tree) + "\n\\end{verbatim}\n}",
			Scale:       0.4,

			CustomStyles: "show id",
		}
		config.Trees = append(config.Trees, TreeConfig)
	}

	config.CustomHead = "\\gtrset{empty name text={}}"
	config.Attribution = fmt.Sprintf("generated at %s", time.Now().Format("2006-01-02"))

	err = renderDocument("templates/document-basic.tex", config, template.FuncMap{}, "test.tex")
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
}

func renderDocument(templateFilename string, data interface{}, customFuncMap template.FuncMap, outputFilename string) error {
	result, err := generations.RenderTemplateFile(templateFilename, data, customFuncMap)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outputFilename, result, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadDatabases(configDatabases []string, databaseBaseDir string) (*generations.MemoryDatabase, error) {
	fmt.Println("loading databases: " + strings.Join(configDatabases, ", "))
	return generations.LoadDatabases(configDatabases, databaseBaseDir)
}
