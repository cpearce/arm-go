package main

import (
	"fmt"
	"testing"
)

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

	iterator := rs.Iterator()
	for iterator.Next() {
		rule := iterator.Get()
		fmt.Println("Rule ", rule)
	}
}
