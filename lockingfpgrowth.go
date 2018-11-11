package main

import (
	"runtime"
	"sync"
)

type taskQueue struct {
	tasks   []masterTask
	mutex   *sync.Mutex
	jobs    int
	condVar *sync.Cond
}

func makeTaskQueue() taskQueue {
	m := sync.Mutex{}
	return taskQueue{
		tasks:   make([]masterTask, 0),
		mutex:   &m,
		jobs:    0,
		condVar: sync.NewCond(&m),
	}
}

func (tq *taskQueue) add(tree *fpTree, itemset []Item, minCount int, counter int) {
	items := frequentItemsInTree(tree, minCount)
	tq.mutex.Lock()
	tq.jobs -= counter
	if len(items) == 0 {
		if tq.jobs == 0 {
			// No other outstanding jobs, wake up any waiters.
			tq.condVar.Broadcast()
		}
		tq.mutex.Unlock()
		return
	}
	mt := masterTask{tree: tree, itemset: itemset, items: items}
	tq.tasks = append(tq.tasks, mt)
	tq.mutex.Unlock()
}

func (tq *taskQueue) nextTask() (*workerTask, bool) {
	tq.mutex.Lock()
	for {
		if len(tq.tasks) != 0 {
			break
		}
		if tq.jobs == 0 {
			tq.mutex.Unlock()
			return nil, false
		}
		tq.condVar.Wait()
	}
	lastTask := &tq.tasks[len(tq.tasks)-1]
	nextWorkerTask := workerTask{
		tree:    lastTask.tree,
		item:    lastTask.items[0],
		itemset: lastTask.itemset,
	}
	if len(lastTask.items) == 1 {
		tq.tasks = tq.tasks[:len(tq.tasks)-1]
	} else {
		lastTask.items = lastTask.items[1:]
	}
	tq.jobs++
	tq.mutex.Unlock()
	return &nextWorkerTask, true
}

func (tq *taskQueue) awaitFinished() {
	tq.mutex.Lock()
	for len(tq.tasks) != 0 || tq.jobs != 0 {
		tq.condVar.Wait()
	}
	tq.mutex.Unlock()
}

func lockingWorker(tq *taskQueue, minCount int, output chan<- itemsetWithCount) {
	for {
		task, ok := tq.nextTask()
		if !ok {
			return
		}
		conditionalTree := makeConditionalTree(task.tree, task.tree.itemList[task.item])
		itemset := appendSorted(task.itemset, task.item)
		output <- itemsetWithCount{
			itemset: itemset,
			count:   conditionalTree.root.count,
		}
		tq.add(conditionalTree, itemset, minCount, 1)
	}
}

func lockingFpGrowth(tree *fpTree, minCount int) []itemsetWithCount {
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

	tq := makeTaskQueue()
	tq.add(tree, []Item{}, minCount, 0)

	for i := 0; i < runtime.NumCPU(); i++ {
		go lockingWorker(&tq, minCount, output)
	}
	tq.awaitFinished()

	close(output)
	return <-c
}
