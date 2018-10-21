package main

import (
	"testing"
)

func itemSliceEquals(a []Item, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ruleEquals(a *Rule, b *Rule) bool {
	return itemSliceEquals(a.Antecedent, b.Antecedent) && itemSliceEquals(a.Consequent, b.Consequent)
}

func TestRules(t *testing.T) {

	ruleData := []Rule{
		NewRule([]Item{1, 2, 3}, []Item{4, 5, 6}, 0.1, 0.2, 0.3),
		NewRule([]Item{1, 2, 3}, []Item{4, 5, 6}, 0.1, 0.2, 0.3),
		NewRule([]Item{1, 2, 3, 4}, []Item{5, 6}, 0.1, 0.2, 0.3),
		NewRule([]Item{1, 2}, []Item{3, 4, 5, 6}, 0.1, 0.2, 0.3),
	}

	rs := NewRuleSet()
	for _, r := range ruleData {
		rs.Insert(&r)
	}

	expectedRules := []Rule{
		NewRule([]Item{1, 2, 3}, []Item{4, 5, 6}, 0.1, 0.2, 0.3),
		NewRule([]Item{1, 2, 3, 4}, []Item{5, 6}, 0.1, 0.2, 0.3),
		NewRule([]Item{1, 2}, []Item{3, 4, 5, 6}, 0.1, 0.2, 0.3),
	}

	foundRule := make([]bool, len(expectedRules))

	iterator := rs.Iterator()
	for iterator.Next() {
		rule := iterator.Get()
		found := false
		for i, r := range expectedRules {
			if ruleEquals(&r, rule) {
				if foundRule[i] {
					t.Errorf("Duplicate find")
				}
				foundRule[i] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Mising rule!")
		}
	}
}
