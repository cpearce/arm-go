package main

import "strings"

// Itemizer converts between a string to an Item type, and vice versa.
type Itemizer struct {
	strToItem map[string]Item
	itemToStr map[Item]string
	numItems  int
}

// Itemize converts a slice of strings to a slice of Items.
func (it *Itemizer) Itemize(values []string) []Item {
	items := make([]Item, len(values))
	j := 0
	it.forEachItem(values, func(i Item) {
		items[j] = i
		j++
	})
	return items[:j]
}

func (it *Itemizer) filter(tokens []string, filter func(Item) bool) []Item {
	items := make([]Item, 0, len(tokens))
	it.forEachItem(tokens, func(i Item) {
		if filter(i) {
			items = append(items, i)
		}
	})
	return items
}

func (it *Itemizer) forEachItem(tokens []string, fn func(Item)) {
	for _, val := range tokens {
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
		fn(itemID)
	}
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
