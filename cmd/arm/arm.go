package main

import (
	"log"
	"time"

	"github.com/cpearce/arm-go/fpgrowth"
	"github.com/pkg/profile"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Println("Association Rule Mining - in Go via FPGrowth")

	args := parseArgsOrDie()
	if args.profile {
		defer profile.Start().Stop()
	}

	log.Println("First pass, counting Item frequencies...")
	start := time.Now()
	ctx, err := fpgrowth.Init(args.input)
	check(err)
	log.Printf("First pass finished in %s", time.Since(start))

	log.Println("Generating frequent itemsets via fpGrowth")
	start = time.Now()
	itemsets, err := ctx.GenerateItemsets(args.minSupport)
	check(err)
	log.Printf("fpGrowth generated %d frequent patterns in %s",
		len(itemsets), time.Since(start))

	if len(args.itemsetsPath) > 0 {
		log.Printf("Writing itemsets to '%s'\n", args.itemsetsPath)
		start := time.Now()
		ctx.WriteItemsets(itemsets, args.itemsetsPath)
		log.Printf(
			"Wrote %d itemsets in %s",
			len(itemsets),
			time.Since(start),
		)
	}

	log.Println("Generating association rules...")
	start = time.Now()
	rules := ctx.GenerateRules(
		itemsets,
		args.minConfidence,
		args.minLift,
	)
	log.Printf(
		"Generated %d association rules in %s",
		len(rules),
		time.Since(start),
	)

	start = time.Now()
	log.Printf("Writing rules to '%s'...", args.output)
	ctx.WriteRules(args.output, rules)
	log.Printf("Wrote %d rules in %s", len(rules), time.Since(start))
}
