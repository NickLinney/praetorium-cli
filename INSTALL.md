# Installation

## Certified release artifact

The pre-alpha release publishes `praetorium-linux-amd64`, an executable that requires no Go, Node.js, Python, or browser runtime on the target host.

Download the binary, `SHA256SUMS`, and `install.sh` from the GitHub release. The installer verifies the checksum and installs the binary; it never compiles source code.

```sh
sh install.sh --version 0.0.0-pre-alpha.1 --base-url https://github.com/NickLinney/praetorium-cli/releases/download/0.0.0-pre-alpha.1
```

By default the installer targets `/usr/local/bin`. Use `--install-dir "$HOME/.local/bin"` for a user-local installation. The installed executable is always named `praetorium`.

## Build locally

```sh
CGO_ENABLED=0 go build -trimpath -o praetorium ./cmd/praetorium
```

For release artifacts, use `scripts/build-release.sh`.

## Coordinator prerequisite

The CLI is only a client. Start or otherwise provide the Praetorium Coordinator at the configured endpoint before executing a prompt. The supplied prototype publishes the Coordinator on `127.0.0.1:8787`.
