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

type Item int

type Itemizer struct {
	strToItem map[string]Item
	itemToStr map[Item]string
	numItems  int
}

func (self *Itemizer) Itemize(values []string) []Item {
	items := make([]Item, 0)
	for _, val := range values {
		val = strings.TrimSpace(val)
		if len(val) == 0 {
			continue
		}
		itemId, found := self.strToItem[val]
		if !found {
			self.numItems++
			itemId = Item(self.numItems)
			self.strToItem[val] = itemId
			self.itemToStr[itemId] = val
		}
		items = append(items, itemId)
	}
	return items
}

func (self *Itemizer) cmp(a Item, b Item) bool {
	return self.itemToStr[a] < self.itemToStr[b]
}

func NewItemizer() Itemizer {
	return Itemizer{
		strToItem: make(map[string]Item),
		itemToStr: make(map[Item]string),
		numItems:  0,
	}
}

func Filter(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

type ItemToNodeSlice map[Item][]*FPNode
type ItemToNode map[Item]*FPNode

type FPNode struct {
	item     Item
	count    int
	parent   *FPNode
	children ItemToNode
}

type FPTree struct {
	root     *FPNode
	itemList ItemToNodeSlice
	counts   map[Item]int
}

const InvalidItem = Item(0)

func NewNode(item Item, parent *FPNode) *FPNode {
	return &FPNode{
		item:     item,
		count:    0,
		parent:   parent,
		children: make(ItemToNode),
	}
}

func NewTree() *FPTree {
	return &FPTree{
		root:     NewNode(InvalidItem, nil),
		itemList: make(ItemToNodeSlice),
		counts:   make(map[Item]int),
	}
}

func (self *FPTree) Insert(transaction []Item, count int) {
	self.root.count += count
	parent := self.root
	for _, item := range transaction {
		node, found := parent.children[item]
		if !found {
			node = NewNode(item, parent)
			parent.children[item] = node
			self.itemList[item] = append(self.itemList[item], node)
		}
		self.counts[item] += count
		node.count += count
		parent = node
	}
}

type ItemSetWithCount struct {
	itemset []Item
	count   int
}

func reverse(a []Item) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}

func isRoot(node *FPNode) bool {
	return node == nil || node.item == InvalidItem
}

func pathFromRootToExcluding(node *FPNode) []Item {
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

func FPGrowth(tree *FPTree, itemset []Item, minCount int) []ItemSetWithCount {
	itemsets := make([]ItemSetWithCount, 0)
	for item, itemList := range tree.itemList {
		if tree.counts[item] < minCount {
			continue
		}
		conditionalTree := NewTree()
		for _, leaf := range itemList {
			transaction := pathFromRootToExcluding(leaf)
			conditionalTree.Insert(transaction, leaf.count)
		}
		path := append(itemset, item)
		itemsets = append(itemsets, ItemSetWithCount{
			itemset: path,
			count:   conditionalTree.root.count,
		})
		x := FPGrowth(conditionalTree, path, minCount)
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

	itemizer := NewItemizer()

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

	minCount := int(math.Floor(minSupport * float64(numTransactions)))

	log.Println("Second pass, building initial tree..")
	start = time.Now()
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	tree := NewTree()
	for scanner.Scan() {
		text := scanner.Text()
		transaction := itemizer.Itemize(strings.Split(text, ","))

		// Strip out items below minCount
		transaction = Filter(transaction, func(i Item) bool {
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

	log.Println("Generating frequent itemsets via FPGrowth")
	start = time.Now()
	itemsWithCount := FPGrowth(tree, make([]Item, 0), minCount)
	log.Printf("FPGrowth generated %d frequent patterns in %s",
		len(itemsWithCount), time.Since(start))

	// Print out frequent itemsets.
	for _, itemWithCount := range itemsWithCount {
		fmt.Println(itemWithCount.itemset, itemWithCount.count)
	}
}
