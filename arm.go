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
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

// Item represents an item.
type Item int

// Itemizer converts between a string to an Item type, and vice versa.
type Itemizer struct {
	strToItem map[string]Item
	itemToStr map[Item]string
	numItems  int
}

// Itemize converts a slice of strings to a slice of Items.
func (it *Itemizer) Itemize(values []string) []Item {
	items := make([]Item, 0)
	for _, val := range values {
		val = strings.TrimSpace(val)
		if len(val) == 0 {
			continue
		}
		itemID, found := it.strToItem[val]
		if !found {
			it.numItems++
			itemID = Item(it.numItems)
			it.strToItem[val] = itemID
			it.itemToStr[itemID] = val
		}
		items = append(items, itemID)
	}
	return items
}

func (it *Itemizer) cmp(a Item, b Item) bool {
	return it.itemToStr[a] < it.itemToStr[b]
}

func newItemizer() Itemizer {
	return Itemizer{
		strToItem: make(map[string]Item),
		itemToStr: make(map[Item]string),
		numItems:  0,
	}
}

func filter(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

type itemToNodeSlice map[Item][]*fpNode
type itemToNode map[Item]*fpNode

type fpNode struct {
	item     Item
	count    int
	parent   *fpNode
	children itemToNode
}

type fpTree struct {
	root     *fpNode
	itemList itemToNodeSlice
	counts   map[Item]int
}

const invalidItem = Item(0)

func newNode(item Item, parent *fpNode) *fpNode {
	return &fpNode{
		item:     item,
		count:    0,
		parent:   parent,
		children: make(itemToNode),
	}
}

func newTree() *fpTree {
	return &fpTree{
		root:     newNode(invalidItem, nil),
		itemList: make(itemToNodeSlice),
		counts:   make(map[Item]int),
	}
}

func (tree *fpTree) Insert(transaction []Item, count int) {
	tree.root.count += count
	parent := tree.root
	for _, item := range transaction {
		node, found := parent.children[item]
		if !found {
			node = newNode(item, parent)
			parent.children[item] = node
			tree.itemList[item] = append(tree.itemList[item], node)
		}
		tree.counts[item] += count
		node.count += count
		parent = node
	}
}

type itemSetWithCount struct {
	itemset []Item
	count   int
}

func reverse(a []Item) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}

func isRoot(node *fpNode) bool {
	return node == nil || node.item == invalidItem
}

func pathFromRootToExcluding(node *fpNode) []Item {
	path := make([]Item, 0)
	for {
		node = node.parent
		if isRoot(node) {
			reverse(path)
			return path
		}
		path = append(path, node.item)
	}
}

func appendSorted(itemset []Item, item Item) []Item {
	xs := make([]Item, len(itemset)+1)
	i := 0
	for idx := range itemset {
		val := itemset[idx]
		if item > val {
			xs[i] = val
			i++
		} else {
			break
		}
	}
	xs[i] = item
	for j := i + 1; j < len(xs); j++ {
		xs[j] = itemset[j-1]
	}
	return xs
}

func fpGrowth(tree *fpTree, itemset []Item, minCount int) []itemSetWithCount {
	itemsets := make([]itemSetWithCount, 0)
	for item, itemList := range tree.itemList {
		if tree.counts[item] < minCount {
			continue
		}
		conditionalTree := newTree()
		for _, leaf := range itemList {
			transaction := pathFromRootToExcluding(leaf)
			conditionalTree.Insert(transaction, leaf.count)
		}
		path := appendSorted(itemset, item)
		itemsets = append(itemsets, itemSetWithCount{
			itemset: path,
			count:   conditionalTree.root.count,
		})
		x := fpGrowth(conditionalTree, path, minCount)
		itemsets = append(itemsets, x...)
	}
	return itemsets
}

func main() {

	log.Println("Association Rule Mining - in Go")
	const minSupport = 0.05

	file, err := os.Open("datasets/kosarak.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	frequency := make(map[Item]int)

	itemizer := newItemizer()

	log.Println("First pass, counting Item frequencies.")
	start := time.Now()
	scanner := bufio.NewScanner(file)
	numTransactions := 0
	for scanner.Scan() {
		numTransactions++
		text := scanner.Text()
		transaction := itemizer.Itemize(strings.Split(text, ","))
		for _, item := range transaction {
			frequency[item]++
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("First pass took %s", time.Since(start))
	log.Printf("Data set contains %d transactions", numTransactions)

	minCount := int(math.Floor(minSupport * float64(numTransactions)))

	log.Println("Second pass, building initial tree..")
	start = time.Now()
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	tree := newTree()
	for scanner.Scan() {
		text := scanner.Text()
		transaction := itemizer.Itemize(strings.Split(text, ","))

		// Strip out items below minCount
		transaction = filter(transaction, func(i Item) bool {
			return frequency[i] >= minCount
		})
		if len(transaction) == 0 {
			continue
		}
		// Sort by decreasing frequency, tie break lexicographically.
		sort.SliceStable(transaction, func(i, j int) bool {
			a := transaction[i]
			b := transaction[j]
			if frequency[a] == frequency[b] {
				return itemizer.cmp(a, b)
			}
			return frequency[a] > frequency[b]
		})

		tree.Insert(transaction, 1)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Building initial tree took %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsWithCount := fpGrowth(tree, make([]Item, 0), minCount)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsWithCount), time.Since(start))

	// Print out frequent itemsets.
	fmt.Println("itemsets:")
	for _, itemWithCount := range itemsWithCount {
		fmt.Println(itemWithCount.itemset, itemWithCount.count)
	}

	start = time.Now()
	rules := generateRules(itemsWithCount, numTransactions, 0.05, 1.5)
	log.Printf("Generated %d association rules in %s",
		len(rules.Rules()), time.Since(start))
	for _, rule := range rules.Rules() {
		fmt.Println(rule)
	}

}
