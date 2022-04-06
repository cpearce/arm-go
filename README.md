# arm-go
[![Build Status](https://travis-ci.org/cpearce/arm-go.svg?branch=master)](https://travis-ci.org/cpearce/arm-go)

An implementation of FPGrowth frequent pattern generation algorithm,
along with association rule generation, in Go.

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