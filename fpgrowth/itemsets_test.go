package fpgrowth

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
		{[]Item{1, 2, 3}, []Item{4, 5, 6}, []Item{1, 2, 3, 4, 5, 6}},
		{[]Item{1}, []Item{1, 2}, []Item{1, 2}},
	}
	test(testCases, union, t)
}

func TestIntersection(t *testing.T) {
	t.Log("TestIntersection")
	testCases := []testCase{
		{[]Item{1}, []Item{}, []Item{}},
		{[]Item{}, []Item{1}, []Item{}},
		{[]Item{1, 2, 3}, []Item{4, 5, 6}, []Item{}},
		{[]Item{1, 2, 3}, []Item{0, 1, 2, 4, 5, 6}, []Item{1, 2}},
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
		{[]Item{148}, 69922},
		{[]Item{11, 148}, 55759},
		{[]Item{6, 11, 148}, 55230},
		{[]Item{148, 218}, 58823},
		{[]Item{11, 148, 218}, 50098},
		{[]Item{6, 11, 148, 218}, 49866},
		{[]Item{6, 148, 218}, 56838},
		{[]Item{6, 148}, 64750},
		{[]Item{218}, 88598},
		{[]Item{6, 218}, 77675},
		{[]Item{11, 218}, 61656},
		{[]Item{6, 11, 218}, 60630},
		{[]Item{3}, 450031},
		{[]Item{3, 6}, 265180},
		{[]Item{1}, 197522},
		{[]Item{1, 3}, 84660},
		{[]Item{1, 3, 6}, 57802},
		{[]Item{1, 6}, 132113},
		{[]Item{1, 11}, 91882},
		{[]Item{1, 6, 11}, 86092},
		{[]Item{6}, 601374},
		{[]Item{4}, 78097},
		{[]Item{27}, 72134},
		{[]Item{6, 27}, 59418},
		{[]Item{7}, 86898},
		{[]Item{7, 11}, 57074},
		{[]Item{6, 7, 11}, 55835},
		{[]Item{6, 7}, 73610},
		{[]Item{11}, 364065},
		{[]Item{6, 11}, 324013},
		{[]Item{3, 11}, 161286},
		{[]Item{3, 6, 11}, 143682},
		{[]Item{55}, 65412},
	}

	input := "../datasets/kosarak.csv"
	itemizer, frequency, numTransactions, err := countItems(input)
	if err != nil {
		t.Error(err)
	}
	itemsets, err := generateFrequentItemsets(
		input,
		0.05,
		itemizer,
		frequency,
		numTransactions,
	)
	if err != nil {
		t.Error(err)
	}
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
		{[]Item{1}, []Item{}, []Item{1}},
		{[]Item{}, []Item{1}, []Item{}},
		{[]Item{1, 2, 3}, []Item{1, 2, 3}, []Item{}},
		{[]Item{1, 2, 3}, []Item{1, 2}, []Item{3}},
		{[]Item{1, 2, 3}, []Item{2}, []Item{1, 3}},
		{[]Item{1, 2, 3}, []Item{3}, []Item{1, 2}},
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
