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
	"log"
	"sort"
	"time"
)

// Rule represents an antecedent implies consequent rule, and stores its
// support, confidence, and lift.
type Rule struct {
	Antecedent []Item
	Consequent []Item
	Support    float64
	Confidence float64
	Lift       float64
}

// NewRule creates a new rule.
func NewRule(antecedent []Item, consequent []Item, support float64, confidence float64, lift float64) Rule {
	return Rule{
		Antecedent: antecedent,
		Consequent: consequent,
		Support:    support,
		Confidence: confidence,
		Lift:       lift,
	}
}

type itemsetWithSupport struct {
	itemset []Item
	support float64
}

func (isl itemsetSupportLookup) Len() int {
	return len(isl.itemsets)
}

func (isl *itemsetSupportLookup) Swap(i, j int) {
	isl.itemsets[i], isl.itemsets[j] = isl.itemsets[j], isl.itemsets[i]
}

func (isl *itemsetSupportLookup) Less(i, j int) bool {
	return itemSliceLess(isl.itemsets[i].itemset, isl.itemsets[j].itemset)
}

type itemsetSupportLookup struct {
	itemsets []itemsetWithSupport
}

func newItemsetSupportLookup() *itemsetSupportLookup {
	return &itemsetSupportLookup{
		itemsets: make([]itemsetWithSupport, 0),
	}
}

func (isl *itemsetSupportLookup) insert(itemset []Item, support float64) {
	isl.itemsets = append(isl.itemsets, itemsetWithSupport{itemset: itemset, support: support})
}

func (isl *itemsetSupportLookup) sort() {
	sort.Sort(isl)
}

func (isl *itemsetSupportLookup) lookup(itemset []Item) float64 {
	idx := sort.Search(len(isl.itemsets), func(idx int) bool {
		return !itemSliceLess(isl.itemsets[idx].itemset, itemset)
	})
	if !itemSliceEquals(isl.itemsets[idx].itemset, itemset) {
		panic("Failed to retrieve itemset support")
	}
	return isl.itemsets[idx].support
}

func createSupportLookup(itemsets []itemsetWithCount, numTransactions int) *itemsetSupportLookup {
	isl := newItemsetSupportLookup()
	f := float64(numTransactions)
	for _, is := range itemsets {
		isl.insert(is.itemset, float64(is.count)/f)
	}
	isl.sort()

	return isl
}

func makeStats(a []Item, c []Item, ac []Item, acSup float64, supportLookup *itemsetSupportLookup) (float64, float64) {
	aSup := supportLookup.lookup(a)
	confidence := acSup / aSup
	cSup := supportLookup.lookup(c)
	lift := acSup / (aSup * cSup)
	return confidence, lift
}

func itemSliceLess(a, b []Item) bool {
	if len(a) < len(b) {
		return true
	} else if len(a) > len(b) {
		return false
	}
	for idx := range a {
		if a[idx] > b[idx] {
			return false
		}
		if a[idx] < b[idx] {
			return true
		}
	}
	return false
}

func sliceOfItemSliceLessThan(slices [][]Item) func(i, j int) bool {
	return func(i, j int) bool {
		return itemSliceLess(slices[i], slices[j])
	}
}

func sortCandidates(candidates [][]Item) {
	sort.SliceStable(candidates, sliceOfItemSliceLessThan(candidates))
}

func prefixMatchLen(a []Item, b []Item) int {
	if len(a) != len(b) {
		panic("prefixMatch called on non-matching length slices")
	}
	for i := range a {
		if a[i] != b[i] {
			return i
		}
	}
	return len(a)
}

func generateRules(itemsets []itemsetWithCount, numTransactions int, minConfidence float64, minLift float64) [][]Rule {
	// Output rules are stored in a slice of slices. As we generate rules, we
	// store them in a slice with capacity `chunkSize`. When the slice fills up,
	// we append it to the output set. If we instead stuck all the rules in a
	// single slice, we'd need to resize the slice as we append more rules, which
	// is slow when we have a lot of rules in the slice.
	output := make([][]Rule, 0)
	const chunkSize int = 10000
	rules := make([]Rule, 0, chunkSize)
	itemsetSupport := createSupportLookup(itemsets, numTransactions)

	lastFeedback := time.Now()

	for index, itemset := range itemsets {
		support := float64(itemset.count) / float64(numTransactions)
		if time.Since(lastFeedback).Seconds() > 20 {
			lastFeedback = time.Now()
			percentComplete := int(float64(index)/float64(countRules(output)+len(rules))*100 + 0.5)
			log.Printf("Progress: %d of %d itemsets processed (%d%%), generated %d rules so far",
				index, len(itemsets), percentComplete, len(rules))
		}
		if len(itemset.itemset) < 2 {
			continue
		}
		// First generation is all possible rules with consequents of size 1.
		candidates := make([][]Item, 0)
		for _, item := range itemset.itemset {
			consequent := []Item{item}
			antecedent := setMinus(itemset.itemset, consequent)
			confidence, lift := makeStats(antecedent, consequent, itemset.itemset, support, itemsetSupport)
			if confidence < minConfidence {
				continue
			}
			if lift >= minLift {
				rules = append(rules, NewRule(antecedent, consequent, support, confidence, lift))
				if len(rules) == chunkSize {
					output = append(output, rules)
					rules = make([]Rule, 0, chunkSize)
				}
			}
			candidates = append(candidates, consequent)
		}
		// Note: candidates should be sorted here.

		// Create subsequent generations by merging consequents which have size-1 items
		// in common in the consequent.
		k := len(itemset.itemset) // size of frequent itemset
		for len(candidates) > 0 && len(candidates[0])+1 < k {
			nextGen := make([][]Item, 0)
			for idx1, c1 := range candidates {
				m := len(c1) // size of consequent.
				for idx2 := idx1 + 1; idx2 < len(candidates); idx2++ {
					c2 := candidates[idx2]
					if prefixMatchLen(c1, c2) != m-1 {
						// The candidates list contains only items of the same length.
						// The candidates list is sorted, and each candidate is sorted.
						// We're trying to merge two consequents which have m-1 items in
						// common. So we can stop searching for c2 once our prefix no
						// longer matches m-1 items, as since the list is sorted, we can't
						// find any more matches after that.
						break
					}

					consequent := union(c1, candidates[idx2])
					antecedent := setMinus(itemset.itemset, consequent)

					confidence, lift := makeStats(antecedent, consequent, itemset.itemset, support, itemsetSupport)
					if confidence < minConfidence {
						continue
					}
					nextGen = append(nextGen, consequent)
					if lift >= minLift {
						rules = append(rules, NewRule(antecedent, consequent, support, confidence, lift))
						if len(rules) == chunkSize {
							output = append(output, rules)
							rules = make([]Rule, 0, chunkSize)
						}
					}
				}
			}
			candidates = nextGen
			sortCandidates(candidates)
		}
	}

	if len(rules) > 0 {
		output = append(output, rules)
	}
	return output
}
