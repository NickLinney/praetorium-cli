# Architecture

## Runtime boundary

Praetorium CLI is an ingress client, not an orchestration engine:

```text
CLI → Praetorium Coordinator → Pi Harness Adapter → Pi Worker → Pi runtime/model
```

The Coordinator's current contract is `POST /execute` with a `HarnessExecutionRequest` and a normalized `HarnessExecutionResult`. The CLI constructs or loads that request, submits it over bounded HTTP, and renders the returned result.

## Package responsibilities

| Package | Responsibility | Explicit negative boundary |
| --- | --- | --- |
| `cmd/praetorium` | command grammar, exit semantics, composition | no direct Pi/model integration |
| `internal/config` | INI serialization and effective precedence | no authority or secret store |
| `internal/request` | prompt normalization and current request validation | no object resolution or attachment ingress |
| `internal/coordinator` | bounded HTTP Coordinator client | no worker-contract construction |
| `internal/render` | human and JSON output | no execution decisions |
| `internal/systemalias` | collision-safe optional shell alias artifact | no shell evaluation or silent overwrite |

## Configuration precedence

```text
explicit CLI option → selected profile → default profile → built-in default
```

INI is only a serialization format. The application operates on typed values after loading and validation.

## Current request semantics

Inline prompts create the minimum request required by the current Coordinator: schema version `0.1.0`, execution/correlation identifiers, `pi.dev`, task text, timeout, and empty deny-by-default context/tool arrays. A supplied request document is validated locally and passed through without inventing identity, IAM, authority, Ordinarium, or ContextPacket resolution.

## Future seams

The current prototype models broader execution channels and future governance gates. The CLI preserves request fields accepted by the Coordinator, but does not claim those future services exist. Attachment support must eventually follow artifact ingress and content-addressed references; it is not implemented by passing local paths to the worker.
