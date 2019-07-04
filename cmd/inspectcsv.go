// Copyright © 2019 Ed Crewe <edmundcrewe@gmail.com>
// Cobra inspectcsv command wrapper
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
//	"github.com/edcrewe/gormcsv/importcsv"
	"github.com/spf13/cobra"
)

// inspectcsvCmd represents the inspectcsv command
var inspectcsvCmd = &cobra.Command{
	Use:   "inspectcsv",
	Short: "Create models from CSV files",
	Long: `The data inspect command. CSV files must be named the same as the target Model / Table`,
	Run: func(cmd *cobra.Command, args []string) {
		Files, _ := cmd.Flags().GetString("files")
		fmt.Printf("Inspect csv for %s\n to generate models.go", Files)
		mcsv := importcsv.ModelCSV{}
		mcsv.PopulateMeta(Files)


	},
}

func init() {
	rootCmd.AddCommand(inspectcsvCmd)
}