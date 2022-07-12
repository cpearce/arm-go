// Copyright 2018 Chris Pearce
// Copyright 2022 Nokia
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
//
// Modified by Nokia into an importable package.
// Modified by Nokia to support custom reader and writer

package arm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

// Item represents an item.
type Item int

type Logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

func writeItemsets(itemsets []itemsetWithCount, itemsetsWriter ItemsetsWriter, itemizer *Itemizer, numTransactions int) error {
	output, err := itemsetsWriter()
	if err != nil {
		return err
	}
	defer output.Close()
	w := bufio.NewWriter(output)
	if _, err := fmt.Fprintln(w, "Itemset,Support"); err != nil {
		return err
	}
	n := float64(numTransactions)
	for _, iwc := range itemsets {
		first := true
		for _, item := range iwc.itemset {
			if !first {
				if _, err := fmt.Fprintf(w, " "); err != nil {
					return err
				}
			}
			first = false
			if _, err := fmt.Fprint(w, itemizer.toStr(item)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, " %f\n", float64(iwc.count)/n); err != nil {
			return err
		}
	}
	return w.Flush()
}

func writeRules(rules [][]Rule, rulesWriter RulesWriter, itemizer *Itemizer) error {
	output, err := rulesWriter()
	if err != nil {
		return err
	}
	defer output.Close()
	w := bufio.NewWriter(output)
	if _, err := fmt.Fprintln(w, "Antecedent => Consequent,Confidence,Lift,Support"); err != nil {
		return err
	}
	for _, chunk := range rules {
		for _, rule := range chunk {
			first := true
			for _, item := range rule.Antecedent {
				if !first {
					if _, err := fmt.Fprintf(w, " "); err != nil {
						return err
					}
				}
				first = false
				if _, err := fmt.Fprint(w, itemizer.toStr(item)); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(w, " => "); err != nil {
				return err
			}
			first = true
			for _, item := range rule.Consequent {
				if !first {
					if _, err := fmt.Fprintf(w, " "); err != nil {
						return err
					}
				}
				first = false
				if _, err := fmt.Fprint(w, itemizer.toStr(item)); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, ",%f,%f,%f\n", rule.Confidence, rule.Lift, rule.Support); err != nil {
				return err
			}
		}
	}
	return w.Flush()
}

func countRules(rules [][]Rule) int {
	n := 0
	for _, chunk := range rules {
		n += len(chunk)
	}
	return n
}

func countItems(itemsReader ItemsReader) (*Itemizer, *itemCount, int, error) {
	file, err := itemsReader()
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
	if err := scanner.Err(); err != nil {
		return nil, nil, 0, err
	}
	return &itemizer, &frequency, numTransactions, nil
}

func generateFrequentItemsets(itemsReader ItemsReader, minSupport float64, itemizer *Itemizer, frequency *itemCount, numTransactions int) ([]itemsetWithCount, error) {
	file, err := itemsReader()
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return fpGrowth(tree, make([]Item, 0), minCount), nil
}

func MineAssociationRules(args Arguments, log Logger) error {
	if err := args.Validate(); err != nil {
		return err
	}

	args_v2 := ArgumentsV2{
		ItemsReader: func() (io.ReadCloser, error) {
			return os.Open(args.Input)
		},
		RulesWriter: func() (io.WriteCloser, error) {
			log.Printf("Writing rules to '%s'...", args.Output)
			return os.Create(args.Output)
		},
		MinSupport:    args.MinSupport,
		MinConfidence: args.MinConfidence,
		MinLift:       args.MinLift,
	}
	if args.ItemsetsPath != "" {
		args_v2.ItemsetsWriter = func() (io.WriteCloser, error) {
			log.Printf("Writing itemsets to '%s'\n", args.ItemsetsPath)
			return os.Create(args.ItemsetsPath)
		}
	}

	return MineAssociationRulesV2(args_v2, log)
}

func MineAssociationRulesV2(args ArgumentsV2, log Logger) error {
	log.Println("Association Rule Mining - in Go via FPGrowth")

	if err := args.Validate(); err != nil {
		return err
	}

	log.Println("First pass, counting Item frequencies...")
	start := time.Now()
	itemizer, frequency, numTransactions, err := countItems(args.ItemsReader)
	if err != nil {
		return err
	}
	log.Printf("First pass finished in %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()

	itemsWithCount, err := generateFrequentItemsets(args.ItemsReader, args.MinSupport, itemizer, frequency, numTransactions)
	if err != nil {
		return err
	}
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsWithCount), time.Since(start))

	if args.ItemsetsWriter != nil {
		start := time.Now()
		writeItemsets(itemsWithCount, args.ItemsetsWriter, itemizer, numTransactions)
		log.Printf("Wrote %d itemsets in %s", len(itemsWithCount), time.Since(start))
	}

	log.Println("Generating association rules...")
	start = time.Now()
	rules := generateRules(itemsWithCount, numTransactions, args.MinConfidence, args.MinLift, log)
	numRules := countRules(rules)
	log.Printf("Generated %d association rules in %s", numRules, time.Since(start))

	start = time.Now()
	writeRules(rules, args.RulesWriter, itemizer)
	log.Printf("Wrote %d rules in %s", numRules, time.Since(start))

	return nil
}
