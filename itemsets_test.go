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
	"testing"
)

type testCase struct {
	a []Item
	b []Item
	c []Item
}

type itemSetOp func([]Item, []Item) []Item

func test(testCases []testCase, f itemSetOp, t *testing.T) {
	for _, tc := range testCases {
		t.Log(tc)
		u := f(tc.a, tc.b)
		if !itemSliceEquals(tc.c, u) {
			t.Error("Result=", u)
		}
	}
}

func TestUnion(t *testing.T) {
	t.Log("TestUnion")
	testCases := []testCase{
		testCase{[]Item{1, 2, 3}, []Item{4, 5, 6}, []Item{1, 2, 3, 4, 5, 6}},
		testCase{[]Item{1}, []Item{1, 2}, []Item{1, 2}},
	}
	test(testCases, union, t)
}

func TestIntersection(t *testing.T) {
	t.Log("TestIntersection")
	testCases := []testCase{
		testCase{[]Item{1}, []Item{}, []Item{}},
		testCase{[]Item{}, []Item{1}, []Item{}},
		testCase{[]Item{1, 2, 3}, []Item{4, 5, 6}, []Item{}},
		testCase{[]Item{1, 2, 3}, []Item{0, 1, 2, 4, 5, 6}, []Item{1, 2}},
	}
	test(testCases, intersection, t)

	for _, tc := range testCases {
		t.Log(tc)
		u := intersectionSize(tc.a, tc.b)
		if u != len(tc.c) {
			t.Error("Result=", u)
		}
	}
}

func containsIWC(expected []itemSetWithCount, observed itemSetWithCount) bool {
	for _, iws := range expected {
		if itemSliceEquals(observed.itemset, iws.itemset) {
			return observed.count == iws.count
		}
	}
	return false
}

func TestFPGrowth(t *testing.T) {
	t.Log("TestFPGrowth")

	expectedItemsets := []itemSetWithCount{
		itemSetWithCount{[]Item{148}, 69922},
		itemSetWithCount{[]Item{11, 148}, 55759},
		itemSetWithCount{[]Item{6, 11, 148}, 55230},
		itemSetWithCount{[]Item{148, 218}, 58823},
		itemSetWithCount{[]Item{11, 148, 218}, 50098},
		itemSetWithCount{[]Item{6, 11, 148, 218}, 49866},
		itemSetWithCount{[]Item{6, 148, 218}, 56838},
		itemSetWithCount{[]Item{6, 148}, 64750},
		itemSetWithCount{[]Item{218}, 88598},
		itemSetWithCount{[]Item{6, 218}, 77675},
		itemSetWithCount{[]Item{11, 218}, 61656},
		itemSetWithCount{[]Item{6, 11, 218}, 60630},
		itemSetWithCount{[]Item{3}, 450031},
		itemSetWithCount{[]Item{3, 6}, 265180},
		itemSetWithCount{[]Item{1}, 197522},
		itemSetWithCount{[]Item{1, 3}, 84660},
		itemSetWithCount{[]Item{1, 3, 6}, 57802},
		itemSetWithCount{[]Item{1, 6}, 132113},
		itemSetWithCount{[]Item{1, 11}, 91882},
		itemSetWithCount{[]Item{1, 6, 11}, 86092},
		itemSetWithCount{[]Item{6}, 601374},
		itemSetWithCount{[]Item{4}, 78097},
		itemSetWithCount{[]Item{27}, 72134},
		itemSetWithCount{[]Item{6, 27}, 59418},
		itemSetWithCount{[]Item{7}, 86898},
		itemSetWithCount{[]Item{7, 11}, 57074},
		itemSetWithCount{[]Item{6, 7, 11}, 55835},
		itemSetWithCount{[]Item{6, 7}, 73610},
		itemSetWithCount{[]Item{11}, 364065},
		itemSetWithCount{[]Item{6, 11}, 324013},
		itemSetWithCount{[]Item{3, 11}, 161286},
		itemSetWithCount{[]Item{3, 6, 11}, 143682},
		itemSetWithCount{[]Item{55}, 65412},
	}

	input := "datasets/kosarak.csv"
	itemizer, frequency, numTransactions := countItems(input)
	itemsets := generateFrequentItemsets(input, 0.05, itemizer, frequency, numTransactions)

	if len(itemsets) != len(expectedItemsets) {
		t.Error("Result=")
	}
	for _, iwc := range itemsets {
		if !containsIWC(expectedItemsets, iwc) {
			t.Error("Generated unexpected itemet")
		}
	}
}
