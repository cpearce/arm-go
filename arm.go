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
)

// Item represents an item.
type Item int

func filter(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func main() {

	log.Println("Association Rule Mining - in Go")
	const minSupport = 0.05

	file, err := os.Open("datasets/kosarak.csv")
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
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("First pass took %s", time.Since(start))
	log.Printf("Data set contains %d transactions", numTransactions)

	minCount := int(math.Floor(minSupport * float64(numTransactions)))

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
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Building initial tree took %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsWithCount := fpGrowth(tree, make([]Item, 0), minCount)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsWithCount), time.Since(start))

	// Print out frequent itemsets.
	fmt.Println("itemsets:")
	for _, itemWithCount := range itemsWithCount {
		fmt.Println(itemWithCount.itemset, itemWithCount.count)
	}

	start = time.Now()
	rules := generateRules(itemsWithCount, numTransactions, 0.05, 1.5)
	log.Printf("Generated %d association rules in %s",
		len(rules.Rules()), time.Since(start))
	for _, rule := range rules.Rules() {
		fmt.Println(rule)
	}

}
