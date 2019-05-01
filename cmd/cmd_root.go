package main

import "github.com/spf13/cobra"

var (
	flagRootOpen            bool
	flagRootVerbose         bool
	flagRootDatabaseBaseDir string
)

func getRootCommand() *cobra.Command {
	var rootCmd = cobra.Command{
		Use:   "generations",
		Short: "generations creates documents from genealogy data",
	}

	rootCmd.AddCommand(getGenealogytreeCommand())

	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&flagRootVerbose, "verbose", "v", true, "verbose output (e.g. lualatex output)")
	flags.StringVarP(&flagRootDatabaseBaseDir, "database-base-dir", "d", "~/.generations/databases/", "Base dir for database file lookup")
	flags.BoolVarP(&flagRootOpen, "open", "o", true, "open generated pdf file")

	return &rootCmd
}
