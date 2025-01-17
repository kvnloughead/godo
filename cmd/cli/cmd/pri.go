package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var priCmd = &cobra.Command{
	Use:   "pri",
	Short: "A brief description of your command",
	Long:  `TODO - add long help text`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pri called")
	},
}

func init() {
	rootCmd.AddCommand(priCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// priCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// priCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
