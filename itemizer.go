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

import "strings"

type itemCount struct {
	counts []int
}

func makeCounts() itemCount {
	return itemCount{counts: make([]int, 0)}
}

func ensureInBounds(slice []int, index int) []int {
	if index < len(slice) {
		return slice
	}
	delta := 1 + index - len(slice)
	return append(slice, make([]int, delta)...)
}

func (ic *itemCount) increment(item Item, count int) {
	idx := int(item)
	ic.counts = ensureInBounds(ic.counts, idx)
	ic.counts[idx] += count
}

func (ic *itemCount) get(item Item) int {
	idx := int(item)
	if idx >= len(ic.counts) {
		return 0
	}
	return ic.counts[idx]
}

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

func (it *Itemizer) toStr(item Item) string {
	s, found := it.itemToStr[item]
	if !found {
		panic("Failed to convert item to string!")
	}
	return s
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
