package main

import "runtime"

type workerTask struct {
	tree    *fpTree
	item    Item
	itemset []Item
}

type masterTask struct {
	tree    *fpTree
	itemset []Item
	items   []Item
}

func master(initialTree *fpTree, minCount int, toWorker chan<- workerTask, fromWorker <-chan masterTask) {
	tasks := make([]masterTask, 0, 100)
	tasks = append(tasks, masterTask{
		tree:    initialTree,
		itemset: []Item{},
		items:   frequentItemsInTree(initialTree, minCount),
	})
	outstandingJobs := 0
	for len(tasks) > 0 || outstandingJobs > 0 {
		if len(tasks) > 0 {
			lastTask := &tasks[len(tasks)-1]
			nextWorkerTask := workerTask{
				tree:    lastTask.tree,
				item:    lastTask.items[0],
				itemset: lastTask.itemset,
			}
			select {
			case task := <-fromWorker:
				if len(task.items) > 0 {
					tasks = append(tasks, task)
				}
				outstandingJobs--
			case toWorker <- nextWorkerTask:
				outstandingJobs++
				if len(lastTask.items) == 1 {
					tasks = tasks[:len(tasks)-1]
				} else {
					lastTask.items = lastTask.items[1:]
				}
			}
		} else {
			task := <-fromWorker
			if len(task.items) > 0 {
				tasks = append(tasks, task)
			}
			outstandingJobs--
		}
	}
	close(toWorker)
}

func frequentItemsInTree(tree *fpTree, minCount int) []Item {
	items := make([]Item, 0, len(tree.itemList))
	for item := range tree.itemList {
		if tree.counts.get(item) > minCount {
			items = append(items, item)
		}
	}
	return items
}

func worker(fromMaster <-chan workerTask, toMaster chan<- masterTask, output chan<- itemsetWithCount, minCount int) {
	for {
		task, ok := <-fromMaster
		if !ok {

			break
		}
		conditionalTree := makeConditionalTree(task.tree, task.tree.itemList[task.item])
		itemset := appendSorted(task.itemset, task.item)
		output <- itemsetWithCount{
			itemset: itemset,
			count:   conditionalTree.root.count,
		}
		items := frequentItemsInTree(conditionalTree, minCount)
		toMaster <- masterTask{tree: conditionalTree, itemset: itemset, items: items}
	}
}

func parallelFpGrowth(tree *fpTree, minCount int) []itemsetWithCount {
	output := make(chan itemsetWithCount)
	c := make(chan []itemsetWithCount, 100000)
	go func() {
		itemsets := make([]itemsetWithCount, 0)
		for iwc := range output {
			itemsets = append(itemsets, iwc)
		}
		c <- itemsets
		close(c)
	}()

	mc := make(chan masterTask, 10000)
	wc := make(chan workerTask, 10000)
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(wc, mc, output, minCount)
	}

	master(tree, minCount, wc, mc)

	close(output)
	return <-c
}
