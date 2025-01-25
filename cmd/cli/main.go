package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "text2promql",
	Short: "Convert natural language to PromQL",
	Long:  `A CLI tool for converting natural language queries to PromQL`,
}

var convertCmd = &cobra.Command{
	Use:   "convert [query]",
	Short: "Convert a natural language query to PromQL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		fmt.Printf("Converting query: %s\n", query)
		// TODO: Add actual conversion logic
		fmt.Println("rate(http_requests_total{status=~\"5..\"}[5m])")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.text2promql.yaml)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
