# jcfg

Enforcing system configuration via json (and ideally scripted w/ jsonnet).

Currently, the golang binary will just be concerned w/ applying a generated
catalog. The generation of the catalog is planned to be handled through jsonnet
scripting (depending how feasible that is), where collections of resources are
defined as functions to be included.

Compiler - whatever it is - should build a catalog by combining a node's facts
with library code. This library code includes module/class/role definitions for
common groupings of resources and advanced resource types. Advanced resource
types should all compile down to a set of fundamental resources (files &
execs) (right? Or do we want plugins with the apply binary?).

## Goals

* Define a system as a 'catalog', a json array of resources to apply.

## Need

* Graph dependencies - apply in parallel? Where graph and ordering defined?
  Easier to just say "compiler", but if jsonnet is compiler we're hosed
  * Apply binary should do it  - ordering defined in catalog, read by apply
    binary
  * OR - hear me out - let // do the work for us. Spin up every loaded resource
    at once, lock dependents until parents have completed or failed
* Generic plugins? Or just apply files and execs?
  * Files and execs *should* work. Will create a few basic modules and see how
    it actually works
* What to do when resource fails? Branching execution?
  * Resource should include `depends` attributes, and applying an (or any
    resource) w/ exit code non-zero should mark all deps as skipped.
  * Or, just have child resources fail when they see a parent has.

## Examples

```
$ go build ./cmd/jcfg
$ JSONNET_PATH=./modules jsonnet ./examples/pkg-config-test.jsonnet > ./pkg-config-test.json

# ./jcfg apply --debug --verbose ./pkg-config-test.json
```


## API

Stolen from kube.
```yaml

api:
kind:
metadata:
  name:
  description:
spec:
ordering:
  before:
  after:
```

## Issues

* Ordering/collectors. Currently only have `afterOk` and `afterFail`. If
  `afterOk` parent failed, we fail (pending `failOk`). What about collecting
resources? Generally users want an ensure, they shouldn't care if the ensure
necessitated a state change
* Ensure present vs absent - newer modules are missing absent cases
* Checks vs sets. FailOk on check means the set is cancelled? Or at least
  hangs, bug
* Control-C to check pending tasks
* Content type secret
