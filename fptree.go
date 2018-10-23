package main

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
	counts   itemCount
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
		counts:   makeCounts(),
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
		tree.counts.increment(item, count)
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
		if tree.counts.get(item) < minCount {
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
