// Copyright 2025 Chris Pearce
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

package fpgrowth

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// Item represents an item.
type Item int

func (fpgctx FpgrowthCtx) WriteItemsets(
	itemsets GeneratedItemsets,
	filePath string,
) error {
	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Itemset,Support")
	n := float64(fpgctx.numTransactions)
	for _, iwc := range itemsets {
		first := true
		for _, item := range iwc.itemset {
			if !first {
				fmt.Fprintf(w, " ")
			}
			first = false
			fmt.Fprint(w, fpgctx.itemizer.toStr(item))
		}
		fmt.Fprintf(w, " %f\n", float64(iwc.count)/n)
	}
	w.Flush()
	return nil
}

func (fpgctx FpgrowthCtx) WriteRules(
	outputPath string,
	rules []Rule,
) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Antecedent => Consequent,Confidence,Lift,Support")
	for _, rule := range rules {
		for i, item := range rule.Antecedent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, fpgctx.itemizer.toStr(item))
		}
		fmt.Fprint(w, " => ")
		for i, item := range rule.Consequent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, fpgctx.itemizer.toStr(item))
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
	return nil
}

func countItems(path string) (*Itemizer, *itemCount, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
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
	if scanner.Err() != nil {
		return nil, nil, 0, scanner.Err()
	}
	return &itemizer, &frequency, numTransactions, nil
}

type GeneratedItemsets []ItemsetWithCount

func (fpgctx FpgrowthCtx) GenerateItemsets(
	minSupport float64,
) (GeneratedItemsets, error) {
	return generateFrequentItemsets(
		fpgctx.inputCsvPath,
		minSupport,
		&fpgctx.itemizer,
		&fpgctx.frequency,
		fpgctx.numTransactions,
	)
}

func generateFrequentItemsets(
	inputCsvPath string,
	minSupport float64,
	itemizer *Itemizer,
	frequency *itemCount,
	numTransactions int,
) ([]ItemsetWithCount, error) {
	file, err := os.Open(inputCsvPath)
	if err != nil {
		return nil, err
	}
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
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return fpGrowth(tree, make([]Item, 0), minCount), nil
}

type FpgrowthCtx struct {
	inputCsvPath    string
	itemizer        Itemizer
	frequency       itemCount
	numTransactions int
}

func Init(inputCsvPath string) (FpgrowthCtx, error) {
	itemizer, frequency, numTransactions, err := countItems(inputCsvPath)
	if err != nil {
		return FpgrowthCtx{}, err
	}
	return FpgrowthCtx{
		inputCsvPath:    inputCsvPath,
		itemizer:        *itemizer,
		frequency:       *frequency,
		numTransactions: numTransactions,
	}, nil
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

func (fpgctx FpgrowthCtx) GenerateRules(
	itemsets GeneratedItemsets,
	minConfidence float64,
	minLift float64,
) []Rule {
	// To avoid expensive resizes when generating an unknown number of rules,
	// generateRules outputs a slice of slices. So merge them together into a
	// single slice to make things cleaner.
	rules2d := generateRules(itemsets, fpgctx.numTransactions, minConfidence, minLift)
	return flatten(rules2d)
}
