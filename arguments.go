// Copyright 2018 Chris Pearce
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strconv"
)

type arguments struct {
	input         string
	output        string
	minSupport    float64
	minConfidence float64
	minLift       float64
	itemsetsPath  string
	profile       bool
}

const usage = `Arguments:
  --input file_path     Input dataset in CSV format.
  --output file_path    File path in which to store output rules. Format:
                        antecedent -> consequent, confidence, lift, support.
  --min-support threshold
                        Minimum itemset support threshold, in range [0,1].
  --min-confidence threshold
                        Minimum rule confidence threshold, in range [0,1].
  --min-lift threshold  Minimum rule lift confidence threshold, in range
                        [1,∞] (optional).
  --itemsets file_path  File path in which to store generated itemsets
                        (optional).
  --profile             Enables profiling via 'profile' package (optional).
`

func parseArgsOrDie() arguments {
	result := arguments{}
	args := os.Args[1:]
	gotMinConf := false
	gotMinSup := false

	minSupErrMsg := "Expected --min-support argument followed by float in range [0,1.0]."
	minConfErrMsg := "Expected --min-confidence argument followed by float in range [0,1.0]."
	minLiftErrMsg := "Expected --min-lift argument followed by float in range [1.0,∞]."

	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(-1)
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--input":
			{
				if i+1 > len(args) {
					fmt.Println("Expected --input to be followed by input CSV path.")
					os.Exit(-1)
				}
				result.input = args[i+1]
				i++
			}
		case "--output":
			{
				if i+1 > len(args) {
					fmt.Println("Expected --output to be followed by output rule path.")
					os.Exit(-1)
				}
				result.output = args[i+1]
				i++
			}
		case "--itemsets":
			{
				if i+1 > len(args) {
					fmt.Println("Expected --itemsets to be followed by output itemsets path.")
					os.Exit(-1)
				}
				result.itemsetsPath = args[i+1]
				i++
			}
		case "--min-support":
			{
				if i+1 > len(args) {
					fmt.Println(minSupErrMsg)
					os.Exit(-1)
				}
				f, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || f < 0.0 || f > 1.0 {
					fmt.Println(minSupErrMsg)
					os.Exit(-1)
				}
				result.minSupport = f
				gotMinSup = true
				i++
			}
		case "--min-confidence":
			{
				if i+1 > len(args) {
					fmt.Println(minConfErrMsg)
					os.Exit(-1)
				}
				f, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || f < 0.0 || f > 1.0 {
					fmt.Println(minConfErrMsg)
					os.Exit(-1)
				}
				result.minConfidence = f
				gotMinConf = true
				i++
			}
		case "--min-lift":
			{
				if i+1 > len(args) {
					fmt.Println(minLiftErrMsg)
					os.Exit(-1)
				}
				f, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || f < 1.0 {
					fmt.Println(minLiftErrMsg)
					os.Exit(-1)
				}
				result.minLift = f
				i++
			}
		case "--profile":
			{
				result.profile = true
			}
		}
	}
	if len(result.input) == 0 {
		fmt.Println("Missing required parameter '--input $csv_path")
		os.Exit(-1)
	}
	if len(result.output) == 0 {
		fmt.Println("Missing required parameter '--output $rule_path")
		os.Exit(-1)
	}
	if !gotMinSup {
		fmt.Println(minSupErrMsg)
		os.Exit(-1)
	}
	if !gotMinConf {
		fmt.Println(minConfErrMsg)
		os.Exit(-1)
	}
	return result
}
