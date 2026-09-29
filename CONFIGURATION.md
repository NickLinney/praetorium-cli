# Configuration

Praetorium uses:

```text
~/.praetorium/config/praetorium-config-default.ini
~/.praetorium/config/praetorium-config-{profile}.ini
```

Set `PRAETORIUM_CONFIG_DIR` for an isolated configuration directory during testing or controlled operation. Profile names are validated before they are used in a path.

## Example

```ini
[cli]
output_format = human

[coordinator]
endpoint = http://127.0.0.1:8787
timeout_seconds = 60

[system_alias]
enabled = false
name = prt
collision_policy = refuse
```

The same keys are available in a named profile. Explicit `--endpoint`, `--timeout`, and `--output` options override profile values.

## Commands

- `config show` displays the effective typed configuration and source paths.
- `config validate` performs validation without mutation.
- `config apply` creates a user-owned alias snippet only when no executable, shell definition, or conflicting managed file is found. It does not edit shell startup files or evaluate configuration as shell code.
- `config status` compares the desired alias with the observed managed alias state.

The optional system alias remains separate from the built-in application aliases `exec` and `prompt`.
