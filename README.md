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

Below example is to look for all method calls (`.*()`) on types named `Client`, which belongs to packages whos import path matches pattern `github.com/tombuildsstuff/giovanni/storage/.*`, among all the packages under `./internal` directory (`./internal/...`):

```shell
terraform-provider-azurerm on  main via 🐹 v1.17.7
💤  ged -p 'github.com/tombuildsstuff/giovanni/storage/.*:Client:.*()' ./internal/...
/home/magodo/github/terraform-provider-azurerm/internal/services/storage/shim/containers_data_plane.go:30:18
/home/magodo/github/terraform-provider-azurerm/internal/services/storage/shim/containers_data_plane.go:53:15
/home/magodo/github/terraform-provider-azurerm/internal/services/storage/shim/containers_data_plane.go:62:19
<omit>
```
