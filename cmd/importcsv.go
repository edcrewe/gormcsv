// Copyright © 2019 Ed Crewe <edmundcrewe@gmail.com>
// Cobra importcsv command wrapper
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/spf13/cobra"
)

// importcsvCmd represents the importcsv command
var importcsvCmd = &cobra.Command{
	Use:   "importcsv",
	Short: "Populates one or more tables from CSV file(s)",
	Long: `The data import command. CSV files must be named the same as the target Model / Table
           or else --model should be supplied`,
	Run: func(cmd *cobra.Command, args []string) {
		Files, _ := cmd.Flags().GetString("files")
		fmt.Printf("Import csv for %s\n", Files)
		mcsv := importcsv.ModelCSV{}
		mcsv.ImportCSV(Files)
	},
}

func init() {
	rootCmd.AddCommand(importcsvCmd)

	// Here you will define your flags and configuration settings specific to importcsv
}
