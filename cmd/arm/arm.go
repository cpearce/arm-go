// Arm generates frequent itemsets and association rules.
//
// Example usage:
//
//	arm \
//	  --input datasets/kosarak.csv \
//	  --output rules.csv \
//	  --itemsets itemsets.csv \
//	  --min-support 0.05 \
//	  --min-confidence 0.05 \
//	  --min-lift 1.5
//
// Command line flags:
//
//   - `input`: path to CSV file containing transactions to analyze. There are some
//     examples in the datasets directory.
//   - `output`: path to file to write the output rules to. Rules are written in CSV
//     format with a header row explaining columns.
//   - `itemsets`: optional path to CSV file to write the generated frequent itemsets
//     to. If specified the large itemsets are written to this file.
//   - `min-support`: minimum support above which itemsets are considered large, and
//     used for rule generation.
//   - `min-confidence`: minimum confidence for rule generation.
//   - `min-lift`: minimum lift for rule generation.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cpearce/arm-go/fpgrowth"
	"github.com/pkg/profile"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Println("Association Rule Mining - in Go via FPGrowth")

	input := flag.String("input", "", "Input dataset in CSV format.")
	output := flag.String("output", "", "File path in which to store output rules. Format: antecedent -> consequent, confidence, lift, support.")
	minSupport := flag.Float64("min-support", 0, "Minimum itemset support threshold, in range [0,1].")
	minConfidence := flag.Float64("min-confidence", 0, "Minimum rule confidence threshold, in range [0,1].")
	minLift := flag.Float64("min-lift", 1, "Minimum rule lift confidence threshold, in range [1,∞] (optional)")
	itemsetsPath := flag.String("itemsets", "", "File path in which to store generated itemsets (optional).")
	enableProfile := flag.Bool("profile", false, "Enables profiling via 'profile' package (optional).")
	flag.Parse()

	if len(*input) == 0 {
		fmt.Println("Missing required parameter '--input $csv_path")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	if len(*output) == 0 {
		fmt.Println("Missing required parameter '--output $rule_path")
		os.Exit(-1)
	}

	if *minSupport < 0.0 || *minSupport > 1.0 {
		fmt.Println("Expected --min-support argument followed by float in range [0,1.0].")
		os.Exit(-1)
	}

	if *minConfidence < 0.0 || *minConfidence > 1.0 {
		fmt.Println("Expected --min-confidence argument followed by float in range [0,1.0].")
		os.Exit(-1)
	}

	if *minLift < 1.0 {
		fmt.Println("Expected --min-lift argument followed by float in range [1.0,∞].")
		os.Exit(-1)
	}

	if *enableProfile {
		defer profile.Start().Stop()
	}

	log.Println("First pass, counting Item frequencies...")
	start := time.Now()
	ctx, err := fpgrowth.Init(*input)
	check(err)
	log.Printf("First pass finished in %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsets, err := ctx.GenerateItemsets(*minSupport)
	check(err)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsets), time.Since(start))

	if len(*itemsetsPath) > 0 {
		log.Printf("Writing itemsets to '%s'\n", *itemsetsPath)
		start := time.Now()
		ctx.WriteItemsets(itemsets, *itemsetsPath)
		log.Printf(
			"Wrote %d itemsets in %s",
			len(itemsets),
			time.Since(start),
		)
	}

	log.Println("Generating association rules...")
	start = time.Now()
	rules := ctx.GenerateRules(
		itemsets,
		*minConfidence,
		*minLift,
	)
	log.Printf(
		"Generated %d association rules in %s",
		len(rules),
		time.Since(start),
	)

	start = time.Now()
	log.Printf("Writing rules to '%s'...", *output)
	ctx.WriteRules(*output, rules)
	log.Printf("Wrote %d rules in %s", len(rules), time.Since(start))
}
