package main

// Rule represents an antecedent implies consequent rult, and stores its
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
	rule        Rule
	hasRule     bool
}

func newRuleTreeNode() *ruleTreeNode {
	return &ruleTreeNode{
		antecedents: make(map[Item]*ruleTreeNode),
		consequents: make(map[Item]*ruleTreeNode),
	}
}

// RuleSet stores a set of rules in a compact tree structure.
type RuleSet struct {
	root *ruleTreeNode
}

// NewRuleSet creates a new RuleSet().
func NewRuleSet() RuleSet {
	return RuleSet{root: newRuleTreeNode()}
}

// Insert inserts a rule into a RuleSet.
func (ruleSet *RuleSet) Insert(rule *Rule) {
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
	parent.hasRule = true
	parent.rule = *rule
}

// RuleSetIterator iterates over a RuleSet.
type RuleSetIterator struct {
	c    <-chan *Rule
	next *Rule
}

// Next attempts to advance the iterator; returns true on success, whereupon
// you can can call Get() to retrieve the value. Returns false when the
// iteration reaches the end.
func (rsi *RuleSetIterator) Next() bool {
	rule, more := <-rsi.c
	rsi.next = rule
	return more
}

// Get retrieves the value at the current point in the iteration.
func (rsi *RuleSetIterator) Get() *Rule {
	return rsi.next
}

func keys(m map[Item]*ruleTreeNode) []Item {
	items := make([]Item, 0)
	for item := range m {
		items = append(items, item)
	}
	return items
}

func traverseRuleTree(node *ruleTreeNode, c chan<- *Rule) {
	for _, n := range node.antecedents {
		traverseRuleTree(n, c)
	}
	for _, n := range node.consequents {
		traverseRuleTree(n, c)
	}
	if node.hasRule {
		c <- &node.rule
	}
}

// Iterator creates a RuleSetIterator which can be used to iterate over
// all Rules in this RuleSet.
func (ruleSet *RuleSet) Iterator() RuleSetIterator {
	c := make(chan *Rule)
	go func() {
		traverseRuleTree(ruleSet.root, c)
		close(c)
	}()
	return RuleSetIterator{c: c}
}
