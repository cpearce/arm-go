// Package fpgrowth implements the FPGrowth algorithm for generating frequent
// itemsets and association rules.
//
// The implementation is highly optimized for performance, converting string
// items into ints, and internally storing itemsets sorted to ensure fast
// access and comparisons.
//
// Input datasets must be in CSV format, without header rows.
//
// To generate association rules, you must create an fpgrowth.Context struct
// by calling the fpgrowth.Init() method. This performs the first pass to count
// the item frequencies, and count the number of transactions. You then call
// fpgrowth.GenerateItemsets() to find the frequent itemsets, and pass that to
// fpgrowth.GenerateRules() to extract the association rules those itemsets
// generate. The itemsets and rules can be written to disk with WriteItemsets()
// and WriteRules() respectively.
package fpgrowth

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// Item represents an item. Use the Itemizer struct to convert back to string
// representation.
type Item int

// WriteItemsets writes itemsets to CSV file.
func (ctx Context) WriteItemsets(
	itemsets GeneratedItemsets,
	filePath string,
) error {
	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Itemset,Support")
	n := float64(ctx.numTransactions)
	for _, iwc := range itemsets {
		first := true
		for _, item := range iwc.Itemset {
			if !first {
				fmt.Fprintf(w, " ")
			}
			first = false
			fmt.Fprint(w, ctx.itemizer.ToStr(item))
		}
		fmt.Fprintf(w, " %f\n", float64(iwc.Count)/n)
	}
	w.Flush()
	return nil
}

// WriteRules writes rules to CSV file.
func (ctx Context) WriteRules(
	outputPath string,
	rules []Rule,
) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(output)
	fmt.Fprintln(w, "Antecedent => Consequent,Confidence,Lift,Support")
	for _, rule := range rules {
		for i, item := range rule.Antecedent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, ctx.itemizer.ToStr(item))
		}
		fmt.Fprint(w, " => ")
		for i, item := range rule.Consequent {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprint(w, ctx.itemizer.ToStr(item))
		}
		fmt.Fprintf(
			w,
			",%f,%f,%f\n",
			rule.Confidence,
			rule.Lift,
			rule.Support,
		)
	}
	w.Flush()
	return nil
}

func countItems(path string) (*Itemizer, *itemCount, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()

	frequency := makeCounts()
	itemizer := newItemizer()

	scanner := bufio.NewScanner(file)
	numTransactions := 0
	for scanner.Scan() {
		numTransactions++
		itemizer.forEachItem(
			strings.Split(scanner.Text(), ","),
			func(item Item) {
				frequency.increment(item, 1)
			})
	}
	if scanner.Err() != nil {
		return nil, nil, 0, scanner.Err()
	}
	return &itemizer, &frequency, numTransactions, nil
}

type GeneratedItemsets []ItemsetWithCount

// GenerateItemsets generates frequent itemsets with support above minSupport.
func (ctx Context) GenerateItemsets(
	minSupport float64,
) (GeneratedItemsets, error) {
	return generateFrequentItemsets(
		ctx.inputCsvPath,
		minSupport,
		&ctx.itemizer,
		&ctx.frequency,
		ctx.numTransactions,
	)
}

func generateFrequentItemsets(
	inputCsvPath string,
	minSupport float64,
	itemizer *Itemizer,
	frequency *itemCount,
	numTransactions int,
) ([]ItemsetWithCount, error) {
	file, err := os.Open(inputCsvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	minCount := max(1, int(math.Ceil(minSupport*float64(numTransactions))))

	scanner := bufio.NewScanner(file)
	tree := newTree()
	for scanner.Scan() {
		transaction := itemizer.filter(
			strings.Split(scanner.Text(), ","),
			func(i Item) bool {
				return frequency.get(i) >= minCount
			})

		if len(transaction) == 0 {
			continue
		}
		// Sort by decreasing frequency, tie break lexicographically.
		sort.SliceStable(transaction, func(i, j int) bool {
			a := transaction[i]
			b := transaction[j]
			if frequency.get(a) == frequency.get(b) {
				return itemizer.cmp(a, b)
			}
			return frequency.get(a) > frequency.get(b)
		})
		tree.Insert(transaction, 1)
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return fpGrowth(tree, make([]Item, 0), minCount), nil
}

// Context stores context for an analysis of itemset transactions.
type Context struct {
	inputCsvPath    string
	itemizer        Itemizer
	frequency       itemCount
	numTransactions int
}

// Init creates a Context. Performs a first pass on dataset, counting
// item frequencies and number of transactions.
func Init(inputCsvPath string) (Context, error) {
	itemizer, frequency, numTransactions, err := countItems(inputCsvPath)
	if err != nil {
		return Context{}, err
	}
	return Context{
		inputCsvPath:    inputCsvPath,
		itemizer:        *itemizer,
		frequency:       *frequency,
		numTransactions: numTransactions,
	}, nil
}

func flatten(rules2d [][]Rule) []Rule {
	n := 0
	for _, r := range rules2d {
		n += len(r)
	}
	rules := make([]Rule, n)
	for _, r := range rules2d {
		rules = append(rules, r...)
	}
	return rules
}

// GenerateRules generates association rules from itemsets with confidence/lift
// above minConfidence/minLift.
func (ctx Context) GenerateRules(
	itemsets GeneratedItemsets,
	minConfidence float64,
	minLift float64,
) []Rule {
	// To avoid expensive resizes when generating an unknown number of rules,
	// generateRules outputs a slice of slices. So merge them together into a
	// single slice to make things cleaner.
	rules2d := generateRules(
		itemsets,
		ctx.numTransactions,
		minConfidence,
		minLift,
	)
	return flatten(rules2d)
}
