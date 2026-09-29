# ADR 0001: Coordinator-only CLI ingress

Status: Accepted for 0.0.0-pre-alpha.1

## Decision

Praetorium CLI submits execution requests only to the Praetorium Coordinator's `/execute` endpoint. It does not call the Pi adapter, Pi worker, Pi runtime, or model provider directly.

## Context

The supplied prototype defines the Coordinator as the request validation, identifier, worker-contract, and normalized-result boundary. Bypassing it would collapse responsibilities and make the CLI incompatible with the current architecture.

## Consequences

The CLI stays small and preserves future adapter/runtime evolution. End-to-end execution depends on a reachable Coordinator, and CLI behavior follows the currently observed `HarnessExecutionRequest`/`HarnessExecutionResult` contract.
