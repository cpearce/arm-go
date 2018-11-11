package main

import (
	"runtime"
	"sync"
)

func fpGrowthOutputToChan(tree *fpTree, itemset []Item, minCount int, output chan<- []itemsetWithCount) {
	results := make([]itemsetWithCount, 0, 1000)
	for item, itemList := range tree.itemList {
		if tree.counts.get(item) < minCount {
			continue
		}
		conditionalTree := makeConditionalTree(tree, itemList)
		path := appendSorted(itemset, item)
		results = append(results, itemsetWithCount{
			itemset: path,
			count:   conditionalTree.root.count,
		})
		fpGrowthOutputToChan(conditionalTree, path, minCount, output)
	}
}

func sparWorker(minCount int, input <-chan sparWorkerTask, output chan<- []itemsetWithCount, wg *sync.WaitGroup) {
	for task := range input {
		results := make([]itemsetWithCount, 0, 1000)
		conditionalTree := makeConditionalTree(task.tree, task.tree.itemList[task.item])
		itemset := []Item{task.item}
		results = append(results, itemsetWithCount{
			itemset: itemset,
			count:   conditionalTree.root.count,
		})
		x := fpGrowth(conditionalTree, itemset, minCount)
		results = append(results, x...)
		output <- results
	}
	wg.Done()
}

type sparWorkerTask struct {
	tree *fpTree
	item Item
}

func sparFpGrowth(tree *fpTree, minCount int) []itemsetWithCount {
	aggregateOutput := make(chan []itemsetWithCount)
	output := make(chan []itemsetWithCount, 100000)
	go func() {
		itemsets := make([]itemsetWithCount, 0, 3082619)
		for iwc := range output {
			itemsets = append(itemsets, iwc...)
		}
		aggregateOutput <- itemsets
		close(aggregateOutput)
	}()

	var wg sync.WaitGroup
	toWorker := make(chan sparWorkerTask, 100000)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go sparWorker(minCount, toWorker, output, &wg)
	}

	items := frequentItemsInTree(tree, minCount)
	for _, item := range items {
		toWorker <- sparWorkerTask{tree: tree, item: item}
	}
	close(toWorker)
	wg.Wait()

	close(output)
	return <-aggregateOutput
}
