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

func writeItemsets(itemsets []itemSetWithCount, outputPath string, itemizer *Itemizer, numTransactions int) {
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
				first = false
			}
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprintf(w, "%f\n", float64(iwc.count)/n)
	}
}

func writeRules(rules RuleSet, outputPath string, itemizer *Itemizer) {
	output, err := os.Create(outputPath)
	check(err)
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Antecedent => Consequent,Confidence,Lift,Support")
	for _, rule := range rules.Rules() {
		first := true
		for _, item := range rule.Antecedent {
			if !first {
				fmt.Fprintf(w, " ")
				first = false
			}
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprint(w, " => ")
		first = true
		for _, item := range rule.Consequent {
			if !first {
				fmt.Fprintf(w, " ")
				first = false
			}
			fmt.Fprint(w, itemizer.toStr(item))
		}
		fmt.Fprintf(w, ",%f,%f,%f\n", rule.Confidence, rule.Lift, rule.Support)
	}
	w.Flush()
}

func main() {
	log.Println("Association Rule Mining - in Go via FPGrowth")

	args := parseArgsOrDie()
	if args.profile {
		defer profile.Start().Stop()
	}

	file, err := os.Open(args.input)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	frequency := makeCounts()
	itemizer := newItemizer()

	log.Println("First pass, counting Item frequencies.")
	start := time.Now()
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
	log.Printf("First pass took %s", time.Since(start))
	log.Printf("Data set contains %d transactions", numTransactions)

	minCount := max(1, int(math.Ceil(args.minSupport*float64(numTransactions))))

	log.Println("Second pass, building initial tree..")
	start = time.Now()
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
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
	log.Printf("Building initial tree took %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsWithCount := fpGrowth(tree, make([]Item, 0), minCount)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsWithCount), time.Since(start))

	if len(args.itemsetsPath) > 0 {
		log.Printf("Writing itemsets to '%s'\n", args.itemsetsPath)
		start = time.Now()
		writeItemsets(itemsWithCount, args.itemsetsPath, &itemizer, numTransactions)
		log.Printf("Wrote %d itemsets in %s", len(itemsWithCount), time.Since(start))
	}

	start = time.Now()
	rules := generateRules(itemsWithCount, numTransactions, args.minConfidence, args.minLift)
	log.Printf("Generated %d association rules in %s",
		len(rules.Rules()), time.Since(start))

	start = time.Now()
	log.Printf("Writing rules to '%s'...", args.output)
	writeRules(rules, args.output, &itemizer)
	log.Printf("Wrote %d rules in %s", rules.Size(), time.Since(start))
}
