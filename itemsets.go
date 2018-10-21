package main

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func intersection(a []Item, b []Item) []Item {
	c := make([]Item, 0, min(len(a), len(b)))
	ap := 0
	bp := 0
	for ap < len(a) && bp < len(b) {
		if a[ap] < b[bp] {
			ap++
		} else if b[bp] < a[ap] {
			bp++
		} else {
			c = append(c, a[ap])
			ap++
			bp++
		}
	}
	return c
}

func union(a []Item, b []Item) []Item {
	c := make([]Item, 0, len(a)+len(b))
	ap := 0
	bp := 0
	for ap < len(a) && bp < len(b) {
		if a[ap] < b[bp] {
			c = append(c, a[ap])
			ap++
		} else if b[bp] < a[ap] {
			c = append(c, b[bp])
			bp++
		} else {
			c = append(c, a[ap])
			ap++
			bp++
		}
	}
	for ap < len(a) {
		c = append(c, a[ap])
		ap++
	}
	for bp < len(b) {
		c = append(c, b[bp])
		bp++
	}
	return c
}

func without(itemset []Item, item Item) ([]Item, []Item) {
	antecedent := make([]Item, 0, len(itemset)-1)
	var consequent []Item
	for idx, it := range itemset {
		if it != item {
			antecedent = append(antecedent, it)
		} else {
			consequent = itemset[idx : idx+1]
		}
	}
	return antecedent, consequent
}

func intersectionSize(a []Item, b []Item) int {
	count := 0
	ap := 0
	bp := 0
	for ap < len(a) && bp < len(b) {
		if a[ap] < b[bp] {
			ap++
		} else if b[bp] < a[ap] {
			bp++
		} else {
			count++
			ap++
			bp++
		}
	}
	return count
}
