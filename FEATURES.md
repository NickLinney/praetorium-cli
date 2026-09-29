# Feature and exit-code catalogue

## Implemented in 0.0.0-pre-alpha.1

| ID | Capability |
| --- | --- |
| F-01 | inline prompt execution |
| F-02 | prompt-file execution |
| F-03 | supplied request-document execution |
| F-04 | stdin prompt execution |
| F-05 | request inspection without execution |
| F-06 | named profile selection |
| F-07 | configuration show/validate/apply/status |
| F-08 | `exec` and `prompt` application aliases |
| F-09 | optional collision-safe `prt`-style system alias artifact |
| F-10 | human and JSON result rendering |
| F-11 | bounded Coordinator HTTP transport |
| F-12 | native, CGO-free Linux build and checksum release artifacts |

## Modeled, not implemented

- richer identity, IAM, authority, and Ordinarium references;
- object/reference resolution and EvidencePacket persistence;
- artifact ingress and ContextPacket materialization;
- future completion contracts, interruption, scheduling, and durable execution state.

## Deferred

Attachments, registries, production credentials, Skills/MCP injection, TUI/GUI/browser UI, persistent sessions, distributed scheduling, automatic updates, `.deb` packaging, shell completion, broad platform certification, and exhaustive performance/SLO testing are outside this MVP.

## Exit codes

| Code | Meaning |
| ---: | --- |
| 0 | successful execution or command |
| 2 | invalid invocation or local request input |
| 3 | invalid/unavailable configuration |
| 4 | Coordinator unreachable or protocol response invalid |
| 5 | Coordinator rejected the request |
| 6 | execution returned failed, timed out, or aborted |
