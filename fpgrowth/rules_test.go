package fpgrowth

import (
	"log"
	"math"
	"testing"
)

func ruleEquals(a *Rule, b *Rule) bool {
	return itemSliceEquals(a.Antecedent, b.Antecedent) &&
		itemSliceEquals(a.Consequent, b.Consequent)
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
	itemsets := []ItemsetWithCount{
		{[]Item{1, 11}, 91882},
		{[]Item{1, 3, 6}, 57802},
		{[]Item{1, 3}, 84660},
		{[]Item{1, 6, 11}, 86092},
		{[]Item{1, 6}, 132113},
		{[]Item{11, 148, 218}, 50098},
		{[]Item{11, 148}, 55759},
		{[]Item{11, 218}, 61656},
		{[]Item{11}, 364065},
		{[]Item{148, 218}, 58823},
		{[]Item{148}, 69922},
		{[]Item{1}, 197522},
		{[]Item{218}, 88598},
		{[]Item{27}, 72134},
		{[]Item{3, 11}, 161286},
		{[]Item{3, 6, 11}, 143682},
		{[]Item{3, 6}, 265180},
		{[]Item{3}, 450031},
		{[]Item{4}, 78097},
		{[]Item{55}, 65412},
		{[]Item{6, 11, 148, 218}, 49866},
		{[]Item{6, 11, 148}, 55230},
		{[]Item{6, 11, 218}, 60630},
		{[]Item{6, 11}, 324013},
		{[]Item{6, 148, 218}, 56838},
		{[]Item{6, 148}, 64750},
		{[]Item{6, 218}, 77675},
		{[]Item{6, 27}, 59418},
		{[]Item{6, 7, 11}, 55835},
		{[]Item{6, 7}, 73610},
		{[]Item{6}, 601374},
		{[]Item{7, 11}, 57074},
		{[]Item{7}, 86898},
	}

	expectedRules := []Rule{
		{[]Item{6}, []Item{1, 11}, 0.0870, 0.143, 1.542},
		{[]Item{11}, []Item{1, 6}, 0.0870, 0.236, 1.772},
		{[]Item{218}, []Item{148}, 0.059, 0.664, 9.400},
		{[]Item{148, 218}, []Item{6}, 0.057, 0.966, 1.591},
		{[]Item{1, 6}, []Item{11}, 0.087, 0.652, 1.772},
		{[]Item{11, 218}, []Item{6, 148}, 0.050, 0.809, 12.366},
		{[]Item{11}, []Item{7}, 0.058, 0.157, 1.786},
		{[]Item{11}, []Item{6, 148, 218}, 0.050, 0.137, 2.386},
		{[]Item{11}, []Item{148, 218}, 0.051, 0.138, 2.316},
		{[]Item{11, 218}, []Item{6}, 0.061, 0.983, 1.619},
		{[]Item{7, 11}, []Item{6}, 0.056, 0.978, 1.610},
		{[]Item{148}, []Item{11}, 0.056, 0.797, 2.168},
		{[]Item{11}, []Item{6, 148}, 0.056, 0.152, 2.319},
		{[]Item{218}, []Item{11}, 0.062, 0.696, 1.892},
		{[]Item{218}, []Item{11, 148}, 0.051, 0.565, 10.040},
		{[]Item{148}, []Item{6}, 0.065, 0.926, 1.524},
		{[]Item{6, 11}, []Item{148}, 0.056, 0.170, 2.413},
		{[]Item{11}, []Item{6, 7}, 0.056, 0.153, 2.063},
		{[]Item{11, 148}, []Item{218}, 0.051, 0.898, 10.040},
		{[]Item{148}, []Item{6, 11, 218}, 0.050, 0.713, 11.645},
		{[]Item{6}, []Item{11, 148, 218}, 0.050, 0.083, 1.639},
		{[]Item{7}, []Item{6, 11}, 0.056, 0.643, 1.963},
		{[]Item{6, 11, 148}, []Item{218}, 0.050, 0.903, 10.089},
		{[]Item{148}, []Item{6, 218}, 0.057, 0.813, 10.360},
		{[]Item{148}, []Item{6, 11}, 0.056, 0.790, 2.413},
		{[]Item{6, 148}, []Item{218}, 0.057, 0.878, 9.809},
		{[]Item{11}, []Item{148}, 0.056, 0.153, 2.168},
		{[]Item{11, 148}, []Item{6}, 0.056, 0.991, 1.631},
		{[]Item{6, 148, 218}, []Item{11}, 0.050, 0.877, 2.386},
		{[]Item{6}, []Item{148, 218}, 0.057, 0.095, 1.591},
		{[]Item{11}, []Item{6, 218}, 0.061, 0.167, 2.123},
		{[]Item{218}, []Item{6, 148}, 0.057, 0.642, 9.809},
		{[]Item{6, 148}, []Item{11}, 0.056, 0.853, 2.319},
		{[]Item{6, 11}, []Item{7}, 0.056, 0.172, 1.963},
		{[]Item{218}, []Item{6, 11, 148}, 0.050, 0.563, 10.089},
		{[]Item{148, 218}, []Item{11}, 0.051, 0.852, 2.316},
		{[]Item{6, 148}, []Item{11, 218}, 0.050, 0.770, 12.366},
		{[]Item{148}, []Item{11, 218}, 0.051, 0.716, 11.504},
		{[]Item{218}, []Item{6, 11}, 0.061, 0.684, 2.091},
		{[]Item{11, 148, 218}, []Item{6}, 0.050, 0.995, 1.639},
		{[]Item{11}, []Item{218}, 0.062, 0.169, 1.892},
		{[]Item{1, 11}, []Item{6}, 0.087, 0.937, 1.542},
		{[]Item{6, 11}, []Item{218}, 0.061, 0.187, 2.091},
		{[]Item{6}, []Item{148}, 0.065, 0.108, 1.524},
		{[]Item{6}, []Item{11, 148}, 0.056, 0.092, 1.631},
		{[]Item{148, 218}, []Item{6, 11}, 0.050, 0.848, 2.590},
		{[]Item{6, 218}, []Item{11}, 0.061, 0.781, 2.123},
		{[]Item{6, 7}, []Item{11}, 0.056, 0.759, 2.063},
		{[]Item{6}, []Item{11, 218}, 0.061, 0.101, 1.619},
		{[]Item{11, 218}, []Item{148}, 0.051, 0.813, 11.504},
		{[]Item{6, 11}, []Item{148, 218}, 0.050, 0.154, 2.590},
		{[]Item{148}, []Item{218}, 0.059, 0.841, 9.400},
		{[]Item{7}, []Item{11}, 0.058, 0.657, 1.786},
		{[]Item{6, 218}, []Item{11, 148}, 0.050, 0.642, 11.398},
		{[]Item{6, 11, 218}, []Item{148}, 0.050, 0.822, 11.645},
		{[]Item{6, 218}, []Item{148}, 0.057, 0.732, 10.360},
		{[]Item{6}, []Item{7, 11}, 0.056, 0.093, 1.610},
		{[]Item{11, 148}, []Item{6, 218}, 0.050, 0.894, 11.398},
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
