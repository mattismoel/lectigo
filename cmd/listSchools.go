/*
Copyright © 2023 Mattis Møl Kristensen <mattismoel@gmail.com>
*/
package cmd

import (
	"log"

	"github.com/mattismoel/lectigo/util"
	"github.com/spf13/cobra"
)

// listSchoolsCmd represents the listSchools command
var listSchoolsCmd = &cobra.Command{
	Use:   "listSchools",
	Short: "Exports the schools registered at Lectio in given encoding scheme at given output path",
	Long: `Exports the schools registered at Lectio in a given encoding scheme. Available schemes are:

json, yaml, xml

The encoded file is exported at the given output path. The path should include at least the base filename. Extension is optional.

Example:

	lego listSchools -f yaml -o ./schoolLists/school.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			log.Fatalf("Could not get format flag: %v\n", err)
		}
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			log.Fatalf("Could not get path flag: %v\n", err)
		}

		err = util.ExportSchools(format, path)
		if err != nil {
			log.Fatalf("Could not export schools list to %v format at path %q: %v\n", format, path, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listSchoolsCmd)

	listSchoolsCmd.Flags().StringP("format", "f", "json", "The format of which the schools list should be exported as")
	listSchoolsCmd.Flags().StringP("path", "o", "./schoolexports/schools", "The path to which the schools list should be exported to")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listSchoolsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listSchoolsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
