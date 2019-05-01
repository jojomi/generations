package main

import (
	"fmt"
	"os"
)

func main() {
	rootCmd := getRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
