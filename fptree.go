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

type itemToNodeSlice map[Item][]*fpNode

type fpNode struct {
	item     Item
	count    int
	parent   *fpNode
	children []*fpNode
	depth    int
}

type fpTree struct {
	root     *fpNode
	itemList itemToNodeSlice
	counts   itemCount
}

const invalidItem = Item(0)

func newNode(item Item, parent *fpNode, depth int) *fpNode {
	return &fpNode{
		item:     item,
		count:    0,
		parent:   parent,
		children: make([]*fpNode, 0),
		depth:    depth,
	}
}

func newTree() *fpTree {
	return &fpTree{
		root:     newNode(invalidItem, nil, 0),
		itemList: make(itemToNodeSlice),
		counts:   makeCounts(),
	}
}

func (tree *fpTree) Insert(transaction []Item, count int) {
	tree.root.count += count
	parent := tree.root
	depth := 1
	for _, item := range transaction {
		var node *fpNode
		for idx := range parent.children {
			if parent.children[idx].item == item {
				node = parent.children[idx]
				break
			}
		}
		if node == nil {
			node = newNode(item, parent, depth)
			parent.children = append(parent.children, node)
			tree.itemList[item] = append(tree.itemList[item], node)
		}
		tree.counts.increment(item, count)
		node.count += count
		parent = node
		depth++
	}
}

type itemsetWithCount struct {
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
	path := make([]Item, 0, node.depth-1)
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

func fpGrowth(tree *fpTree, itemset []Item, minCount int) []itemsetWithCount {
	itemsets := make([]itemsetWithCount, 0)
	for item, itemList := range tree.itemList {
		if tree.counts.get(item) < minCount {
			continue
		}
		conditionalTree := newTree()
		for _, leaf := range itemList {
			transaction := pathFromRootToExcluding(leaf)
			conditionalTree.Insert(transaction, leaf.count)
		}
		path := appendSorted(itemset, item)
		itemsets = append(itemsets, itemsetWithCount{
			itemset: path,
			count:   conditionalTree.root.count,
		})
		x := fpGrowth(conditionalTree, path, minCount)
		itemsets = append(itemsets, x...)
	}
	return itemsets
}
