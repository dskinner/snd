# Package snd [![GoDoc](https://godoc.org/dasa.cc/snd?status.svg)](https://godoc.org/dasa.cc/snd)

Package snd provides methods and types for sound processing and synthesis.

```
go get dasa.cc/snd
```

## Tests

In addition to regular unit tests, there are plot tests that produce images
saved to a plots/ folder. This depends on package gonum/plot and requires an
additional tag flag to enable as follows:

```
go get github.com/gonum/plot
go test -tags plot
```
