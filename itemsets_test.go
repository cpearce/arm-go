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
