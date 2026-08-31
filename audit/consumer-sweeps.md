# Round-3 consumer sweeps

Recorded 2026-08-31 against isomorphic engine commit
`f058cb9a615c67a4c056f527dd249ac7cb4611f8`.

Each consumer was checked from a clean local clone at the recorded source
commit. The following battery passed for every consumer:

```text
iso build
iso conform --lane gen
iso conform --lane core
iso conform --lane walk
```

| consumer | source commit | result |
|---|---|---|
| Hypercubic-AI/app-carddemo | `533dcf9ef7351e339f2f6436d60a8a1924e01f5e` | GREEN |
| Hypercubic-AI/app-pin | `f95c89908fa63bf02018edb6a45d1b5bb908daf9` | GREEN |
| Hypercubic-AI/app-renal-pricer | `4075a11b0dcebada7185d5827a1e66146c3f9d73` | GREEN |

The PIN build reported its pre-existing deterministic duplicate-stem warnings
for `INCLOTES`, `INCUBPRO`, and `ITPINSAP`; no lane failed and no consumer tree
was modified to suppress them. The core lane in all three consumers included
runtime conformance, Dafny verification smoke, and resolve-all.
