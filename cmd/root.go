// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "analog",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		dumpFileName, _ := cmd.Flags().GetString("dump-file")
		dfDNHeader, _ := cmd.Flags().GetString("df-dn-header")
		dfHPHeader, _ := cmd.Flags().GetString("df-hp-header")

		assessmentFileName, _ := cmd.Flags().GetString("assessment-file")
		afDNHeader, _ := cmd.Flags().GetString("af-dn-header")
		afHPHeader, _ := cmd.Flags().GetString("af-hp-header")

		outputFile, _ := cmd.Flags().GetString("output-file")

		dumpSplit := strings.Split(dumpFileName, ".")

		if len(dumpSplit) == 1 {
			return errors.New("--dump-file must be file with csv extension")
		}

		if dumpSplit[len(dumpSplit)-1] != "csv" {
			return errors.New("--dump-file is not type csv")
		}

		assessmentSplit := strings.Split(assessmentFileName, ".")

		if len(assessmentSplit) == 1 {
			return errors.New("--assessment-file must be file with csv extension")
		}

		if assessmentSplit[len(assessmentSplit)-1] != "csv" {
			return errors.New("--assessment-file is not type csv")
		}

		dumpFile, err := os.OpenFile(dumpFileName, os.O_RDONLY, os.ModePerm)

		if err != nil {
			return err
		}

		defer dumpFile.Close()

		assessmentFile, err := os.OpenFile(assessmentFileName, os.O_RDONLY, os.ModePerm)

		if err != nil {
			return err
		}

		defer assessmentFile.Close()

		assessmentReader := csv.NewReader(assessmentFile)
		// numOfAssessmentFields := assessmentReader.FieldsPerRecord

		// if numOfAssessmentFields < afDNCol {
		// 	return errors.New("--af-dn-col is greater than total columns")
		// }
		// if numOfAssessmentFields < afHPCol {
		// 	return errors.New("--af-hp-col is greater than total columns")
		// }

		dumpReader := csv.NewReader(dumpFile)
		// numOfDumpFields := dumpReader.FieldsPerRecord

		// if numOfDumpFields < dfDNCol {
		// 	return errors.New("--df-dn-col is greater than total columns")
		// }

		// if numOfDumpFields < dfHPCol {
		// 	return errors.New("--df-hp-col is greater than total columns")
		// }

		dumpMap := make(map[string]string)
		isFirstRow := true
		dumpHPHeaderNum := -1
		dumpDNHeaderNum := -1

		// Reading dump file and getting house pairs and dns
		for {
			columns, err := dumpReader.Read()

			if err != nil {
				if err == io.EOF {
					break
				}

				return errors.New(err.Error())
			}

			if isFirstRow {
				for i, v := range columns {
					if strings.ToLower(v) == strings.ToLower(dfDNHeader) {
						dumpDNHeaderNum = i
					}
					if strings.ToLower(v) == strings.ToLower(dfHPHeader) {
						dumpHPHeaderNum = i
					}
				}

				if dumpDNHeaderNum == -1 {
					return errors.New("Could not find header name " + dfDNHeader + " in dump file")
				}
				if dumpHPHeaderNum == -1 {
					return errors.New("Could not find header name " + dfHPHeader + " in dump file")
				}

				isFirstRow = false
			} else {
				housePair := columns[dumpHPHeaderNum]
				dn := columns[dumpDNHeaderNum]

				dumpMap[dn] = housePair
			}
		}

		outFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, os.ModePerm)

		if err != nil {
			return err
		}

		defer outFile.Close()

		outFileReader := bufio.NewReader(outFile)

		isFirstRow = true

		assessmentHPHeaderNum := -1
		assessmentDNHeaderNum := -1
		assessmentMap := make(map[string]string)

		// Read assessment file and find compare dns to find house pairs
		for {
			columns, err := assessmentReader.Read()

			if err != nil {
				if err == io.EOF {
					break
				}

				return errors.New(err.Error())
			}

			if isFirstRow {
				for i, v := range columns {
					if strings.ToLower(v) == strings.ToLower(afDNHeader) {
						assessmentDNHeaderNum = i
					}
					if strings.ToLower(v) == strings.ToLower(afHPHeader) {
						assessmentHPHeaderNum = i
					}
				}

				if assessmentDNHeaderNum == -1 {
					return errors.New("Could not find header name " + afDNHeader + " in assessment file")
				}
				if assessmentHPHeaderNum == -1 {
					return errors.New("Could not find header name " + afHPHeader + " in assessment file")
				}

				isFirstRow = false
			} else {
				// If assessment sheet has something in house pair, don't overwrite
				if columns[assessmentHPHeaderNum] == "" {
					for dn, hp := range dumpMap {
						if len(dn) > columns[assessmentDNHeaderNum] {
							if strings.Contains(dn, assessmentDNHeaderNum) {

							}
						}
					}
				}
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.analog.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("dump-file", "d", "", "Dump file which is exported generally from old analog system")
	rootCmd.MarkFlagRequired("dump-file")

	rootCmd.Flags().StringP("df-dn-header", "", "", "Dump file direct number column number.  Ex: If direct number is in column 'F' then column number would be 6")
	rootCmd.MarkFlagRequired("df-dn-header")

	rootCmd.Flags().StringP("df-hp-header", "", "", "Dump file house pair column number.  Ex: If house pair is in column 'H' then column number would be 8")
	rootCmd.MarkFlagRequired("df-hp-header")

	rootCmd.Flags().StringP("assessment-file", "a", "", "Assessment file where analog assement has been done for facility")
	rootCmd.MarkFlagRequired("assessment-file")

	rootCmd.Flags().StringP("af-dn-header", "", "", "Assessment file direct number column number.  Ex: If direct number is in column 'F' then column number would be 6")
	rootCmd.MarkFlagRequired("af-dn-header")

	rootCmd.Flags().StringP(
		"af-hp-header",
		"",
		"",
		`
		Assessment file house pair column number.  
		This column is where the cross references between the values of --dump-file and --assessment-file will be dumped.
		Ex: If direct number is in column 'F' then column number would be 6
		`,
	)
	rootCmd.MarkFlagRequired("af-hp-header")

	rootCmd.Flags().StringP("output-file", "", "", "New file that will combine all the files")
	rootCmd.MarkFlagRequired("output-file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".analog" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".analog")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
