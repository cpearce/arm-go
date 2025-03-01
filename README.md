# arm-go

![Build Status](https://github.com/cpearce/arm-go/actions/workflows/go.yml/badge.svg)

An implementation of FPGrowth frequent pattern generation algorithm,
along with association rule generation, in Go.

Implementation is heavily optimized for performance.

This finds relationships of the form "people who buy X also buy Y",
and also determines the strengths (confidence, lift, support) of those
relationships.

For an overview of assocation rule mining,
see Chapter 5 of Introduction to Data Mining, Kumar et al:
[Association Analysis: Basic Concepts and Algorithms](https://www-users.cs.umn.edu/~kumar001/dmbook/ch5_association_analysis.pdf).

## The `arm` command line tool

To build, first you must [download and install Go](https://golang.org/dl/).
Then install with:

```
go install github.com/cpearce/arm-go/cmd/arm
```

That will download, build, and install the `arm` binary into your `$GOPATH/bin`
directory.

To run the binary, ensure `$GOPATH/bin` is in your path, and run:

```
arm \
  --input datasets/kosarak.csv \
  --output rules.csv \
  --itemsets itemsets.csv \
  --min-support 0.05 \
  --min-confidence 0.05 \
  --min-lift 1.5
```

Command line flags:

* `input`: path to CSV file containing transactions to analyze. There are some
examples in the [datasets/](datasets/) directory.
* `output`: path to file to write the output rules to. Rules are written in CSV
format with a header row explaining columns.
* `itemsets`: optional path to CSV file to write the generated frequent itemsets
to. If specified the large itemsets are written to this file.
* `min-support`: minimum support above which itemsets are considered large, and
used for rule generation.
* `min-confidence`: minimum confidence for rule generation.
* `min-lift`: minimum lift for rule generation.

## The `fpgrowth` package

The underlying implementation can be used as a library as well. See the go docs
for more information.

## Development

To run the cmd/arm binary:

```
go run github.com/cpearce/arm-go/cmd/arm \
  --input datasets/kosarak.csv \
  --output rules.csv \
  --itemsets itemsets.csv \
  --min-support 0.05 \
  --min-confidence 0.05 \
  --min-lift 1.5
```

Unit tests:

```
  $ go test ./...
```

