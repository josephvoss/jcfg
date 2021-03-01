# jcfg

Enforcing system configuration via json (and ideally scripted w/ jsonnet).

Currently, this will just be concerned w/ *applying* a generated catalog. The
*generation* of the catalog will be handled later, either through creating a
server-side/compiler binary, or just saying "use jsonnet".

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
* Generic plugins? Or just apply files and execs?
  * Files and execs should work
* What to do when resource fails? Branching execution?
  * Resource should include `depends` attributes, and applying an (or any
    resource) w/ exit code non-zero should mark all deps as skipped.

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
