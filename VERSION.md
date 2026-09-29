# Praetorium CLI 0.0.0-pre-alpha.1

Lifecycle: Pre-Alpha snapshot

This release proves the first usable vertical slice: a native Go CLI can load configuration, construct or inspect a current `HarnessExecutionRequest`, submit it to the Praetorium Coordinator, render the normalized result, and return stable nonzero failures. It also publishes a CGO-free Debian amd64 binary, installer, and checksum manifest.

Compatibility guarantees are not provided. Breaking changes remain expected in pre-alpha.
