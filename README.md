# Natural Cruise

Isomorphic application fixture for Software AG's Apache-2.0 Natural cruise
samples. The app combines the interactive `NTCRUISE` library with the headless
`CRUISE16` library. Exact upstream commits and assembly paths live in
`audit/upstream-sources.json`.

`RDCRUISE` is retained as audited source but excluded from the v1 compile set;
its missing upstream `PROCESS PAGE` target is recorded in
`audit/dispositions.json`.
