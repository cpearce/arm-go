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
	"math"
	"testing"
)

func ruleEquals(a *Rule, b *Rule) bool {
	return itemSliceEquals(a.Antecedent, b.Antecedent) && itemSliceEquals(a.Consequent, b.Consequent)
}

func TestWithout(t *testing.T) {
	a, c := without([]Item{1, 2, 3}, Item(2))
	if !itemSliceEquals(a, []Item{1, 3}) || !itemSliceEquals(c, []Item{2}) {
		t.Error()
	}
}

func find(rules [][]Rule, r *Rule) (*Rule, bool) {
	for _, chunk := range rules {
		for _, v := range chunk {
			if ruleEquals(r, &v) {
				return &v, true
			}
		}
	}
	return nil, false
}

func TestGenerateRules(t *testing.T) {
	// itemsets generated for kosarak with minsup 0.05.
	itemsets := []itemsetWithCount{
		itemsetWithCount{[]Item{1, 11}, 91882},
		itemsetWithCount{[]Item{1, 3, 6}, 57802},
		itemsetWithCount{[]Item{1, 3}, 84660},
		itemsetWithCount{[]Item{1, 6, 11}, 86092},
		itemsetWithCount{[]Item{1, 6}, 132113},
		itemsetWithCount{[]Item{11, 148, 218}, 50098},
		itemsetWithCount{[]Item{11, 148}, 55759},
		itemsetWithCount{[]Item{11, 218}, 61656},
		itemsetWithCount{[]Item{11}, 364065},
		itemsetWithCount{[]Item{148, 218}, 58823},
		itemsetWithCount{[]Item{148}, 69922},
		itemsetWithCount{[]Item{1}, 197522},
		itemsetWithCount{[]Item{218}, 88598},
		itemsetWithCount{[]Item{27}, 72134},
		itemsetWithCount{[]Item{3, 11}, 161286},
		itemsetWithCount{[]Item{3, 6, 11}, 143682},
		itemsetWithCount{[]Item{3, 6}, 265180},
		itemsetWithCount{[]Item{3}, 450031},
		itemsetWithCount{[]Item{4}, 78097},
		itemsetWithCount{[]Item{55}, 65412},
		itemsetWithCount{[]Item{6, 11, 148, 218}, 49866},
		itemsetWithCount{[]Item{6, 11, 148}, 55230},
		itemsetWithCount{[]Item{6, 11, 218}, 60630},
		itemsetWithCount{[]Item{6, 11}, 324013},
		itemsetWithCount{[]Item{6, 148, 218}, 56838},
		itemsetWithCount{[]Item{6, 148}, 64750},
		itemsetWithCount{[]Item{6, 218}, 77675},
		itemsetWithCount{[]Item{6, 27}, 59418},
		itemsetWithCount{[]Item{6, 7, 11}, 55835},
		itemsetWithCount{[]Item{6, 7}, 73610},
		itemsetWithCount{[]Item{6}, 601374},
		itemsetWithCount{[]Item{7, 11}, 57074},
		itemsetWithCount{[]Item{7}, 86898},
	}

	expectedRules := []Rule{
		Rule{[]Item{6}, []Item{1, 11}, 0.0870, 0.143, 1.542},
		Rule{[]Item{11}, []Item{1, 6}, 0.0870, 0.236, 1.772},
		Rule{[]Item{218}, []Item{148}, 0.059, 0.664, 9.400},
		Rule{[]Item{148, 218}, []Item{6}, 0.057, 0.966, 1.591},
		Rule{[]Item{1, 6}, []Item{11}, 0.087, 0.652, 1.772},
		Rule{[]Item{11, 218}, []Item{6, 148}, 0.050, 0.809, 12.366},
		Rule{[]Item{11}, []Item{7}, 0.058, 0.157, 1.786},
		Rule{[]Item{11}, []Item{6, 148, 218}, 0.050, 0.137, 2.386},
		Rule{[]Item{11}, []Item{148, 218}, 0.051, 0.138, 2.316},
		Rule{[]Item{11, 218}, []Item{6}, 0.061, 0.983, 1.619},
		Rule{[]Item{7, 11}, []Item{6}, 0.056, 0.978, 1.610},
		Rule{[]Item{148}, []Item{11}, 0.056, 0.797, 2.168},
		Rule{[]Item{11}, []Item{6, 148}, 0.056, 0.152, 2.319},
		Rule{[]Item{218}, []Item{11}, 0.062, 0.696, 1.892},
		Rule{[]Item{218}, []Item{11, 148}, 0.051, 0.565, 10.040},
		Rule{[]Item{148}, []Item{6}, 0.065, 0.926, 1.524},
		Rule{[]Item{6, 11}, []Item{148}, 0.056, 0.170, 2.413},
		Rule{[]Item{11}, []Item{6, 7}, 0.056, 0.153, 2.063},
		Rule{[]Item{11, 148}, []Item{218}, 0.051, 0.898, 10.040},
		Rule{[]Item{148}, []Item{6, 11, 218}, 0.050, 0.713, 11.645},
		Rule{[]Item{6}, []Item{11, 148, 218}, 0.050, 0.083, 1.639},
		Rule{[]Item{7}, []Item{6, 11}, 0.056, 0.643, 1.963},
		Rule{[]Item{6, 11, 148}, []Item{218}, 0.050, 0.903, 10.089},
		Rule{[]Item{148}, []Item{6, 218}, 0.057, 0.813, 10.360},
		Rule{[]Item{148}, []Item{6, 11}, 0.056, 0.790, 2.413},
		Rule{[]Item{6, 148}, []Item{218}, 0.057, 0.878, 9.809},
		Rule{[]Item{11}, []Item{148}, 0.056, 0.153, 2.168},
		Rule{[]Item{11, 148}, []Item{6}, 0.056, 0.991, 1.631},
		Rule{[]Item{6, 148, 218}, []Item{11}, 0.050, 0.877, 2.386},
		Rule{[]Item{6}, []Item{148, 218}, 0.057, 0.095, 1.591},
		Rule{[]Item{11}, []Item{6, 218}, 0.061, 0.167, 2.123},
		Rule{[]Item{218}, []Item{6, 148}, 0.057, 0.642, 9.809},
		Rule{[]Item{6, 148}, []Item{11}, 0.056, 0.853, 2.319},
		Rule{[]Item{6, 11}, []Item{7}, 0.056, 0.172, 1.963},
		Rule{[]Item{218}, []Item{6, 11, 148}, 0.050, 0.563, 10.089},
		Rule{[]Item{148, 218}, []Item{11}, 0.051, 0.852, 2.316},
		Rule{[]Item{6, 148}, []Item{11, 218}, 0.050, 0.770, 12.366},
		Rule{[]Item{148}, []Item{11, 218}, 0.051, 0.716, 11.504},
		Rule{[]Item{218}, []Item{6, 11}, 0.061, 0.684, 2.091},
		Rule{[]Item{11, 148, 218}, []Item{6}, 0.050, 0.995, 1.639},
		Rule{[]Item{11}, []Item{218}, 0.062, 0.169, 1.892},
		Rule{[]Item{1, 11}, []Item{6}, 0.087, 0.937, 1.542},
		Rule{[]Item{6, 11}, []Item{218}, 0.061, 0.187, 2.091},
		Rule{[]Item{6}, []Item{148}, 0.065, 0.108, 1.524},
		Rule{[]Item{6}, []Item{11, 148}, 0.056, 0.092, 1.631},
		Rule{[]Item{148, 218}, []Item{6, 11}, 0.050, 0.848, 2.590},
		Rule{[]Item{6, 218}, []Item{11}, 0.061, 0.781, 2.123},
		Rule{[]Item{6, 7}, []Item{11}, 0.056, 0.759, 2.063},
		Rule{[]Item{6}, []Item{11, 218}, 0.061, 0.101, 1.619},
		Rule{[]Item{11, 218}, []Item{148}, 0.051, 0.813, 11.504},
		Rule{[]Item{6, 11}, []Item{148, 218}, 0.050, 0.154, 2.590},
		Rule{[]Item{148}, []Item{218}, 0.059, 0.841, 9.400},
		Rule{[]Item{7}, []Item{11}, 0.058, 0.657, 1.786},
		Rule{[]Item{6, 218}, []Item{11, 148}, 0.050, 0.642, 11.398},
		Rule{[]Item{6, 11, 218}, []Item{148}, 0.050, 0.822, 11.645},
		Rule{[]Item{6, 218}, []Item{148}, 0.057, 0.732, 10.360},
		Rule{[]Item{6}, []Item{7, 11}, 0.056, 0.093, 1.610},
		Rule{[]Item{11, 148}, []Item{6, 218}, 0.050, 0.894, 11.398},
	}

	rules := generateRules(itemsets, 990002, 0.05, 1.5)
	log.Printf("Generated %d rules", len(rules))
	for _, rule := range rules {
		log.Print(rule)
	}
	if countRules(rules) != len(expectedRules) {
		t.Error("Incorrect number of rules generated")
	}

	for _, expected := range expectedRules {
		r, found := find(rules, &expected)
		if !found {
			t.Error("expected rule not found, ", expected)
			continue
		}
		if math.Abs(r.Support-expected.Support) > .001 {
			t.Error("Support doesn't match for ", r)
		}
		if math.Abs(r.Confidence-expected.Confidence) > .001 {
			t.Error("Confidence doesn't match for ", r)
		}
		if math.Abs(r.Lift-expected.Lift) > .001 {
			t.Error("Lift doesn't match for ", r, " expected ", expected)
		}
	}
}
