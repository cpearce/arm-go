package main

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

type ruleTreeNode struct {
	antecedents map[Item]*ruleTreeNode
	consequents map[Item]*ruleTreeNode
	hasRule     bool
	index       int
}

func newRuleTreeNode() *ruleTreeNode {
	return &ruleTreeNode{
		antecedents: make(map[Item]*ruleTreeNode),
		consequents: make(map[Item]*ruleTreeNode),
	}
}

// RuleSet stores a set of rules in a compact tree structure.
type RuleSet struct {
	root  *ruleTreeNode
	rules []Rule
}

// NewRuleSet creates a new RuleSet().
func NewRuleSet() RuleSet {
	return RuleSet{root: newRuleTreeNode()}
}

// Insert inserts a rule into a RuleSet.
func (ruleSet *RuleSet) Insert(rule Rule) {
	parent := ruleSet.root
	for _, item := range rule.Antecedent {
		node, found := parent.antecedents[item]
		if !found {
			node = newRuleTreeNode()
			parent.antecedents[item] = node
		}
		parent = node
	}
	for _, item := range rule.Consequent {
		node, found := parent.consequents[item]
		if !found {
			node = newRuleTreeNode()
			parent.consequents[item] = node
		}
		parent = node
	}
	if !parent.hasRule {
		ruleSet.rules = append(ruleSet.rules, rule)
		parent.hasRule = true
		parent.index = len(ruleSet.rules) - 1
	}
}

// Size returns the number of rules in the set.
func (ruleSet *RuleSet) Size() int {
	return len(ruleSet.rules)
}

// Rules returns the set of rules.
func (ruleSet *RuleSet) Rules() []Rule {
	return ruleSet.rules
}

// Get returns (rule,true) if this RuleSet contains the rule, (nil,false)
// otherwise.
func (ruleSet *RuleSet) Get(rule *Rule) (*Rule, bool) {
	parent := ruleSet.root
	for _, item := range rule.Antecedent {
		node, found := parent.antecedents[item]
		if !found {
			return nil, false
		}
		parent = node
	}
	for _, item := range rule.Consequent {
		node, found := parent.consequents[item]
		if !found {
			return nil, false
		}
		parent = node
	}
	if !parent.hasRule {
		return nil, false
	}
	return &ruleSet.rules[parent.index], true
}

type itemsetSupportLookup struct {
	children map[Item]*itemsetSupportLookup
	support  float64
}

func newItemsetSupportLookup() *itemsetSupportLookup {
	return &itemsetSupportLookup{
		children: make(map[Item]*itemsetSupportLookup),
	}
}

func (isl *itemsetSupportLookup) insert(itemset []Item, support float64) {
	parent := isl
	for _, item := range itemset {
		node, found := parent.children[item]
		if !found {
			node = newItemsetSupportLookup()
			parent.children[item] = node
		}
		parent = node
	}
	if parent.support != 0.0 {
		panic("Duplicate insertion")
	}
	parent.support = support
}

func (isl *itemsetSupportLookup) lookup(itemset []Item) float64 {
	parent := isl
	for _, item := range itemset {
		node, found := parent.children[item]
		if !found {
			panic("Lookup of itemset not in itemsetSupportLookup!")
		}
		parent = node
	}
	return parent.support
}

func createSupportLookup(itemsets []itemSetWithCount, numTransactions int) *itemsetSupportLookup {
	isl := newItemsetSupportLookup()
	f := float64(numTransactions)
	for _, is := range itemsets {
		isl.insert(is.itemset, float64(is.count)/f)
	}

	return isl
}

func makeStats(a []Item, c []Item, supportLookup *itemsetSupportLookup) (float64, float64, float64) {
	ac := union(a, c)
	acSup := supportLookup.lookup(ac)
	aSup := supportLookup.lookup(a)
	confidence := acSup / aSup
	cSup := supportLookup.lookup(c)
	lift := acSup / (aSup * cSup)
	return acSup, confidence, lift
}

func generateRules(itemsets []itemSetWithCount, numTransactions int, minConfidence float64, minLift float64) RuleSet {
	output := NewRuleSet()
	itemsetSupport := createSupportLookup(itemsets, numTransactions)

	for _, itemset := range itemsets {
		if len(itemset.itemset) < 2 {
			continue
		}
		// First generation is all possible rules with consequents of size 1.
		candidates := NewRuleSet()
		for _, item := range itemset.itemset {
			a, c := without(itemset.itemset, item)
			support, confidence, lift := makeStats(a, c, itemsetSupport)
			if confidence < minConfidence {
				continue
			}
			candidates.Insert(NewRule(a, c, support, confidence, lift))
		}
		// Create subsequent generations by merging rules which have size-1 items
		// in common in the consequent.
		for len(candidates.Rules()) > 0 {
			rules := candidates.Rules()
			nextGen := NewRuleSet()
			for idx1, r1 := range rules {
				for idx2 := idx1 + 1; idx2 < len(rules); idx2++ {
					r2 := rules[idx2]
					if len(r1.Consequent) != len(r2.Consequent) {
						continue
					}
					if intersectionSize(r1.Consequent, r2.Consequent) != len(r1.Consequent)-1 {
						continue
					}
					antecedent := intersection(r1.Antecedent, r2.Antecedent)
					if len(antecedent) == 0 {
						continue
					}
					consequent := union(r1.Consequent, r2.Consequent)
					if len(consequent) == 0 {
						continue
					}
					support, confidence, lift := makeStats(antecedent, consequent, itemsetSupport)
					if confidence < minConfidence {
						continue
					}
					nextGen.Insert(NewRule(antecedent, consequent, support, confidence, lift))
				}
			}
			for _, rule := range candidates.Rules() {
				if rule.Lift >= minLift {
					output.Insert(rule)
				}
			}
			candidates = nextGen
		}
	}

	return output
}
