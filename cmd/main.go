package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
		addGlobals(&treeConfig, config)
		database := generations.NewMemoryDatabase()

		for _, db := range treeConfig.Databases {
			err := database.ParseYamlFile(db)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		o := treeConfig.RenderTreeOptions
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

	err = renderDocument(config.Template, config, "test.tex")
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
}

func addGlobals(treeConfig *TreeConfig, config Config) {
	if len(treeConfig.Databases) == 0 {
		treeConfig.Databases = config.Databases
	}
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
