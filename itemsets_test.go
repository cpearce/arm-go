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

func containsIWC(expected []ItemsetWithCount, observed ItemsetWithCount) bool {
	for _, iws := range expected {
		if itemSliceEquals(observed.itemset, iws.itemset) {
			return observed.count == iws.count
		}
	}
	return false
}

func TestFPGrowth(t *testing.T) {
	t.Log("TestFPGrowth")

	expectedItemsets := []ItemsetWithCount{
		ItemsetWithCount{[]Item{148}, 69922},
		ItemsetWithCount{[]Item{11, 148}, 55759},
		ItemsetWithCount{[]Item{6, 11, 148}, 55230},
		ItemsetWithCount{[]Item{148, 218}, 58823},
		ItemsetWithCount{[]Item{11, 148, 218}, 50098},
		ItemsetWithCount{[]Item{6, 11, 148, 218}, 49866},
		ItemsetWithCount{[]Item{6, 148, 218}, 56838},
		ItemsetWithCount{[]Item{6, 148}, 64750},
		ItemsetWithCount{[]Item{218}, 88598},
		ItemsetWithCount{[]Item{6, 218}, 77675},
		ItemsetWithCount{[]Item{11, 218}, 61656},
		ItemsetWithCount{[]Item{6, 11, 218}, 60630},
		ItemsetWithCount{[]Item{3}, 450031},
		ItemsetWithCount{[]Item{3, 6}, 265180},
		ItemsetWithCount{[]Item{1}, 197522},
		ItemsetWithCount{[]Item{1, 3}, 84660},
		ItemsetWithCount{[]Item{1, 3, 6}, 57802},
		ItemsetWithCount{[]Item{1, 6}, 132113},
		ItemsetWithCount{[]Item{1, 11}, 91882},
		ItemsetWithCount{[]Item{1, 6, 11}, 86092},
		ItemsetWithCount{[]Item{6}, 601374},
		ItemsetWithCount{[]Item{4}, 78097},
		ItemsetWithCount{[]Item{27}, 72134},
		ItemsetWithCount{[]Item{6, 27}, 59418},
		ItemsetWithCount{[]Item{7}, 86898},
		ItemsetWithCount{[]Item{7, 11}, 57074},
		ItemsetWithCount{[]Item{6, 7, 11}, 55835},
		ItemsetWithCount{[]Item{6, 7}, 73610},
		ItemsetWithCount{[]Item{11}, 364065},
		ItemsetWithCount{[]Item{6, 11}, 324013},
		ItemsetWithCount{[]Item{3, 11}, 161286},
		ItemsetWithCount{[]Item{3, 6, 11}, 143682},
		ItemsetWithCount{[]Item{55}, 65412},
	}

	input := "datasets/kosarak.csv"
	itemizer, frequency, numTransactions := countItems(input)
	itemsets := generateFrequentItemsets(
		input,
		0.05,
		itemizer,
		frequency,
		numTransactions,
	)

	if len(itemsets) != len(expectedItemsets) {
		t.Error("Result=")
	}
	for _, iwc := range itemsets {
		if !containsIWC(expectedItemsets, iwc) {
			t.Error("Generated unexpected itemet")
		}
	}
}

func TestSetMinus(t *testing.T) {
	t.Log("TestSetMinus")
	testCases := []testCase{
		testCase{[]Item{1}, []Item{}, []Item{1}},
		testCase{[]Item{}, []Item{1}, []Item{}},
		testCase{[]Item{1, 2, 3}, []Item{1, 2, 3}, []Item{}},
		testCase{[]Item{1, 2, 3}, []Item{1, 2}, []Item{3}},
		testCase{[]Item{1, 2, 3}, []Item{2}, []Item{1, 3}},
		testCase{[]Item{1, 2, 3}, []Item{3}, []Item{1, 2}},
	}
	for _, test := range testCases {
		c := setMinus(test.a, test.b)
		if !itemSliceEquals(c, test.c) {
			t.Error(
				"Fail: ",
				test.a,
				" minus ",
				test.b,
				" should be ",
				test.c,
				" got ",
				c,
			)
		}
	}
}
