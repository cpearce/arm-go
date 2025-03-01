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
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pkg/profile"
)

// Item represents an item.
type Item int

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeItemsets(
	itemsets []ItemsetWithCount,
	outputPath string,
	itemizer *Itemizer,
	numTransactions int,
) {
	output, err := os.Create(outputPath)
	check(err)
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Itemset,Support")
	n := float64(numTransactions)
	for _, iwc := range itemsets {
		first := true
		for _, item := range iwc.itemset {
			if !first {
				fmt.Fprintf(w, " ")
			}
			first = false
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprintf(w, " %f\n", float64(iwc.count)/n)
	}
	w.Flush()
}

func writeRules(rules []Rule, outputPath string, itemizer *Itemizer) {
	output, err := os.Create(outputPath)
	check(err)
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Antecedent => Consequent,Confidence,Lift,Support")
	for _, rule := range rules {
		for i, item := range rule.Antecedent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprint(w, " => ")
		for i, item := range rule.Consequent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprintf(
			w,
			",%f,%f,%f\n",
			rule.Confidence,
			rule.Lift,
			rule.Support,
		)
	}
	w.Flush()
}

func countRules(rules [][]Rule) int {
	n := 0
	for _, chunk := range rules {
		n += len(chunk)
	}
	return n
}

func countItems(path string) (*Itemizer, *itemCount, int) {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	frequency := makeCounts()
	itemizer := newItemizer()

	scanner := bufio.NewScanner(file)
	numTransactions := 0
	for scanner.Scan() {
		numTransactions++
		itemizer.forEachItem(
			strings.Split(scanner.Text(), ","),
			func(item Item) {
				frequency.increment(item, 1)
			})
	}
	check(scanner.Err())
	return &itemizer, &frequency, numTransactions
}

func generateFrequentItemsets(
	path string,
	minSupport float64,
	itemizer *Itemizer,
	frequency *itemCount,
	numTransactions int,
) []ItemsetWithCount {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	minCount := max(1, int(math.Ceil(minSupport*float64(numTransactions))))

	scanner := bufio.NewScanner(file)
	tree := newTree()
	for scanner.Scan() {
		transaction := itemizer.filter(
			strings.Split(scanner.Text(), ","),
			func(i Item) bool {
				return frequency.get(i) >= minCount
			})

		if len(transaction) == 0 {
			continue
		}
		// Sort by decreasing frequency, tie break lexicographically.
		sort.SliceStable(transaction, func(i, j int) bool {
			a := transaction[i]
			b := transaction[j]
			if frequency.get(a) == frequency.get(b) {
				return itemizer.cmp(a, b)
			}
			return frequency.get(a) > frequency.get(b)
		})
		tree.Insert(transaction, 1)
	}
	check(scanner.Err())

	return fpGrowth(tree, make([]Item, 0), minCount)
}

type FpgrowthCtx struct {
	inputCsvPath    string
	itemizer        Itemizer
	frequency       itemCount
	numTransactions int
}

func Init(inputCsvPath string) FpgrowthCtx {
	itemizer, frequency, numTransactions := countItems(inputCsvPath)
	return FpgrowthCtx{
		inputCsvPath:    inputCsvPath,
		itemizer:        *itemizer,
		frequency:       *frequency,
		numTransactions: numTransactions,
	}
}

type GeneratedItemsets []ItemsetWithCount

func (fpgctx FpgrowthCtx) GenerateItemsets(
	minSupport float64,
) GeneratedItemsets {
	return generateFrequentItemsets(
		fpgctx.inputCsvPath,
		minSupport,
		&fpgctx.itemizer,
		&fpgctx.frequency,
		fpgctx.numTransactions,
	)
}

func (fpgctx FpgrowthCtx) WriteItemsets(
	itemsets GeneratedItemsets,
	filePath string,
) {
	writeItemsets(
		itemsets,
		filePath,
		&fpgctx.itemizer,
		fpgctx.numTransactions,
	)
}

func flatten(rules2d [][]Rule) []Rule {
	n := 0
	for _, r := range rules2d {
		n += len(r)
	}
	rules := make([]Rule, n)
	for _, r := range rules2d {
		rules = append(rules, r...)
	}
	return rules
}

func (fpctx FpgrowthCtx) GenerateRules(
	itemsets GeneratedItemsets,
	minConfidence float64,
	minLift float64,
) []Rule {
	// To avoid expensive resizes when generating an unknown number of rules,
	// generateRules outputs a slice of slices. So merge them together into a
	// single slice to make things cleaner.
	rules2d := generateRules(itemsets, fpctx.numTransactions, minConfidence, minLift)
	return flatten(rules2d)
}

func (fpctx FpgrowthCtx) WriteRules(
	outputPath string,
	rules []Rule,
) {
	writeRules(rules, outputPath, &fpctx.itemizer)
}

func main() {
	log.Println("Association Rule Mining - in Go via FPGrowth")

	args := parseArgsOrDie()
	if args.profile {
		defer profile.Start().Stop()
	}

	log.Println("First pass, counting Item frequencies...")
	start := time.Now()
	ctx := Init(args.input)
	log.Printf("First pass finished in %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsets := ctx.GenerateItemsets(args.minSupport)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsets), time.Since(start))

	if len(args.itemsetsPath) > 0 {
		log.Printf("Writing itemsets to '%s'\n", args.itemsetsPath)
		start := time.Now()
		ctx.WriteItemsets(itemsets, args.itemsetsPath)
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
		args.minConfidence,
		args.minLift,
	)
	log.Printf(
		"Generated %d association rules in %s",
		len(rules),
		time.Since(start),
	)

	start = time.Now()
	log.Printf("Writing rules to '%s'...", args.output)
	ctx.WriteRules(args.output, rules)
	log.Printf("Wrote %d rules in %s", len(rules), time.Since(start))
}
