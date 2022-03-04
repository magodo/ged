# ged

![workflow](https://github.com/magodo/ged/actions/workflows/go.yml/badge.svg)

**G**et **e**xternal **d**ependencies of Go code.

In case of finding external functions/methods, the tool currently only support statically dispatched calls.

## Install

    go install github.com/magodo/ged

## Usage


    ged [options] [packages]

    Options:
      -p string
            <pkg path>:<ident>[:[<field>|<method>()]]

## Example

Below example is to look for all method calls of `.Position()` on type `FileSet`, which is defined in package `go/token`, among the `ged` codebase:

```shell
ged on  main via 🐹 v1.17.7 
💤  ged -p 'go/token:FileSet:Position()' .
go/token FileSet.Position():
        /home/magodo/github/ged/pattern.go:98:7
        /home/magodo/github/ged/pattern.go:128:21
        /home/magodo/github/ged/pattern.go:167:20
```
