package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/jojomi/go-script"
	"github.com/jojomi/go-spew/spew"
	"github.com/jojomi/strtpl"

	"github.com/jojomi/generations"
	"github.com/spf13/cobra"
)

var (
	flagRootConfigFile string
	flagRootShowConfig bool
	flagRootOpen       bool
	flagRootCheckIDs   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "generations",
		Short: "generations creates genealogytrees using (Lua)LaTeX and the awesome genealogytree package",
		Run:   commandRoot,
	}
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&flagRootConfigFile, "config-file", "c", "config/document.yml", "config filename")
	flags.BoolVarP(&flagRootShowConfig, "debug-config", "d", false, "show parsed config")
	flags.BoolVarP(&flagRootOpen, "open", "o", true, "open generated pdf file")
	flags.BoolVarP(&flagRootCheckIDs, "check-ids", "i", true, "error on unlinked IDs")
	rootCmd.AddCommand(getTestCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func commandRoot(c *cobra.Command, args []string) {
	sc := script.NewContext()

	if !sc.FileExists(flagRootConfigFile) {
		log.Fatalf("file not found: %s\n", flagRootConfigFile)
	}

	var config Config
	data, err := ioutil.ReadFile(flagRootConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	config.SetDefaults()
	setDefaultOutputPath(&config, flagRootConfigFile)

	if flagRootShowConfig {
		spew.Dump(config)
		os.Exit(1)
	}

	templateData := map[string]interface{}{
		"Now": time.Now(),
	}
	config.Title = strtpl.MustEval(config.Title, templateData)
	config.Attribution = strtpl.MustEval(config.Attribution, templateData)
	config.Date = strtpl.MustEval(config.Date, templateData)

	// load config, use it
	for i, treeConfig := range config.Trees {
		treeConfig.AddGlobals(config)
		database := generations.NewMemoryDatabase()

		for _, db := range treeConfig.Databases {
			err := database.ParseYamlFile(db)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		o := treeConfig.RenderTreeOptions
		o.FailForIDLookup = flagRootCheckIDs
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
		if o.RenderPersonOptions != nil {
			o.RenderPersonOptions.TemplateFilename = o.TemplateFilenamePerson
		}

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
			Config     Config
			TreeConfig TreeConfig
			Options    map[string]interface{}
		}{
			Config:     config,
			TreeConfig: treeConfig,
			Options:    treeConfig.Templates.Tree.Options,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(6)
		}
		renderedTrees = renderedTrees + string(renderedTree)
	}
	config.RenderedTrees = string(renderedTrees)

	err = os.MkdirAll(filepath.Dir(config.OutputFilename), 0750)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	renderConfig := struct {
		Config  Config
		Options map[string]interface{}
	}{
		config,
		config.Templates.Document.Options,
	}
	err = renderDocument(config.Templates.Document.Filename, renderConfig, strings.Replace(config.OutputFilename, ".pdf", ".tex", -1))
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	}

	err = compileDocument(
		strings.Replace(config.OutputFilename, ".pdf", ".tex", -1),
		2,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = minifyDocument(
		config.OutputFilename,
		strings.Replace(config.OutputFilename, ".pdf", ".small.pdf", -1),
	)
	if err != nil {
		log.Fatal(err)
	}

	if flagRootOpen {
		openDocument(strings.Replace(config.OutputFilename, ".pdf", ".small.pdf", -1))
	}
}

func compileDocument(inputFile string, numRuns int) error {
	sc := script.NewContext()
	if !sc.CommandExists("lualatex") {
		return fmt.Errorf("lualatex not found in PATH")
	}
	dir := filepath.Dir(inputFile)

	localCommand := script.LocalCommandFrom("lualatex --interaction=nonstopmode --shell-escape")
	localCommand.AddAll(
		"--aux-directory="+dir,
		"--output-directory="+dir,
		inputFile,
	)
	for i := 0; i < numRuns; i++ {
		pr, err := sc.ExecuteDebug(localCommand)
		if err != nil {
			return err
		}
		if !pr.Successful() {
			return fmt.Errorf("lualatex compilation not successful")
		}
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
	sc := script.NewContext()
	if !sc.CommandExists("xdg-open") {
		return
	}

	command := strtpl.MustEval(
		`xdg-open "{{ .Filename }}"`,
		struct {
			Filename string
		}{
			Filename: filename,
		},
	)
	localCommand := script.LocalCommandFrom(command)
	_, _ = sc.ExecuteDebug(localCommand)
}

func setDefaultOutputPath(config *Config, filename string) {
	if config.OutputFilename != "" {
		return
	}

	f := filepath.Base(filename)
	f = strings.TrimSuffix(f, ".yml")
	config.OutputFilename = filepath.Join("output", f+".pdf")
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
	var config Config

	files, err := ioutil.ReadDir(filepath.Join("..", "testdata", "database"))
	if err != nil {
		log.Fatal(err)
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

	err = renderDocument("templates/document-basic.tex", config, "test.tex")
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
}

func renderDocument(templateFilename string, data interface{}, outputFilename string) error {
	result, err := generations.RenderTemplateFile(templateFilename, data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outputFilename, result, 0644)
	if err != nil {
		return err
	}
	return nil
}
