# arm-go
An implementation of FPGrowth frequent pattern generation algorithm,
along with association rule generation, in Go.

Original code by Chris Pearce, https://github.com/cpearce/arm-go.

Modified by Nokia into an importable package and to support custom
reader and writer.

This finds relationships of the form "people who buy X also buy Y",
and also determines the strengths (confidence, lift, support) of those
relationships.

For an overview of assocation rule mining,
see Chapter 5 of Introduction to Data Mining, Kumar et al:
[Association Analysis: Basic Concepts and Algorithms](https://www-users.cs.umn.edu/~kumar001/dmbook/ch5_association_analysis.pdf).

To build, [download and install Go](https://golang.org/dl/) and clone this
repository, and build with:
```
  $ go build ./cmd/arm-go
```
You can then run from the command line, for example:
```
  $ ./arm-go --input datasets/kosarak.csv \
             --output rules \
             --itemsets itemsets \
             --min-support 0.05 \
             --min-confidence 0.05 \
             --min-lift 1.5
```
To run unit tests:
```
  $ go test
```

To use as a library, import `github.com/nokia/arm-go`,
set up `arm.Arguments`, and call `arm.MineAssociationRules`.
For example:
```go
package main

import (
    "log"

    "github.com/nokia/arm-go"
)

func main() {
    args := arm.Arguments{
        Input:         "datasets/kosarak.csv",
        Output:        "rules",
        MinSupport:    0.05,
        MinConfidence: 0.05,
        MinLift:       1.5,
        ItemsetsPath:  "itemsets",
    }
    if err := arm.MineAssociationRules(args, log.Default()); err != nil {
        panic(err)
    }
}
```

Or by using custom readers and writers. For example:
```go
package main

import (
	"io"
	"log"
	"os"

	"github.com/nokia/arm-go"
)

func main() {
	args := arm.ArgumentsV2{
		ItemsReader: func() (io.ReadCloser, error) {
			return os.Open("datasets/kosarak.csv")
		},
		RulesWriter: func() (io.WriteCloser, error) {
			return os.Create("rules")
		},
		ItemsetsWriter: func() (io.WriteCloser, error) {
			return os.Create("itemsets")
		},
		MinSupport:    0.05,
		MinConfidence: 0.05,
		MinLift:       1.5,
	}
	if err := arm.MineAssociationRulesV2(args, log.Default()); err != nil {
		panic(err)
	}
}
```
