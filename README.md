# arm-go

![Build Status](https://github.com/cpearce/arm-go/actions/workflows/go.yml/badge.svg)

An implementation of FPGrowth frequent pattern generation algorithm,
along with association rule generation, in Go.

This finds relationships of the form "people who buy X also buy Y",
and also determines the strengths (confidence, lift, support) of those
relationships.

For an overview of assocation rule mining,
see Chapter 5 of Introduction to Data Mining, Kumar et al:
[Association Analysis: Basic Concepts and Algorithms](https://www-users.cs.umn.edu/~kumar001/dmbook/ch5_association_analysis.pdf).

To build, [download and install Go](https://golang.org/dl/) and clone this
repository to $GO_PATH/src/arm-go, and build with:
```
  $ go build arm-go
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