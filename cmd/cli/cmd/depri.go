/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// depriCmd represents the depri command
var depriCmd = &cobra.Command{
	Use:   "depri",
	Short: "A brief description of your command",
	Long:  `TODO - add long help text`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("depri called")
	},
}

func init() {
	rootCmd.AddCommand(depriCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// depriCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// depriCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
