# Package snd [![GoDoc](https://godoc.org/dasa.cc/snd?status.svg)](https://godoc.org/dasa.cc/snd)

Package snd provides methods and types for sound processing and synthesis.

```
go get dasa.cc/snd
```

## Windows

Tested with msys2. Additional setup required due to how gomobile currently links to openal on windows. This should go away in the future stepping away from gomobile's exp/al package.

```
pacman -S mingw-w64-x86_64-openal
cd /mingw64/lib
cp libopenal.a libOpenAL32.a
cp libopenal.dll.a libOpenAL32.dll.a
```

## Tests

In addition to regular unit tests, there are plot tests that produce images
saved to a plots/ folder. This depends on package gonum/plot and requires an
additional tag flag to enable as follows:

```
go get github.com/gonum/plot
go test -tags plot
```

## SndObj

This package was very much inspired by Victor Lazzarini's [SndObj Library](http://sndobj.sourceforge.net/)
for which I spent countless hours enjoying, and without it I may never have come to program.
