# Praetorium CLI

Praetorium CLI is a small, native Go client for the Praetorium Coordinator. The first release is **0.0.0-pre-alpha.1** and targets Debian GNU/Linux 13 on amd64.

The supported execution path is:

```text
terminal → praetorium → Coordinator /execute → Harness Adapter → Worker → Pi → normalized result
```

The CLI deliberately calls only the Coordinator. It does not call Pi, the adapter, a model provider, or future authority/resource services directly.

## Quick start

```sh
./praetorium version
./praetorium config show
./praetorium execute --prompt "Reply with one short sentence."
```

The Coordinator defaults to `http://127.0.0.1:8787`. Configure it with `~/.praetorium/config/praetorium-config-default.ini` or override it for one invocation:

```sh
./praetorium --endpoint http://127.0.0.1:8787 execute --prompt "Hello"
```

See [INSTALL.md](INSTALL.md) and [CONFIGURATION.md](CONFIGURATION.md) for installation and profile details.

## Commands

- `praetorium execute --prompt TEXT`
- `praetorium execute --prompt-file PATH`
- `praetorium execute --request PATH`
- `praetorium inspect --request PATH`
- `praetorium exec ...` and `praetorium prompt TEXT`
- `praetorium config show|validate|apply|status`
- `praetorium version`

Use `--output json` for the Coordinator's normalized result. Exit codes are documented in [FEATURES.md](FEATURES.md).

## Status boundary

Implemented capabilities are listed in [FEATURES.md](FEATURES.md). Object registries, attachments, IAM, authority resolution, Skills/MCP injection, persistent sessions, schedulers, `.deb` packaging, and broad Linux certification remain deferred by design.
