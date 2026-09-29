# Release process

1. Confirm scope, tests, documentation, and a clean working tree.
2. Update `VERSION.md`, `CHANGELOG.md`, and applicable README/status text.
3. Build the release artifacts with `scripts/build-release.sh`.
4. Verify the binary, installer, checksum, and smoke behavior.
5. Promote the feature-complete `pre-alpha/1` branch to `main`.
6. Create annotated tag `0.0.0-pre-alpha.1` on the authoritative released commit.
7. Push the branch and tag, then publish the GitHub prerelease from that tag.
8. Record exact commit, tag, artifact hashes, tests, and any governance qualification in the external Project release record.

`main` is release-authoritative; `dev` and `test` are integration/validation surfaces; `pre-alpha/1` is the release vehicle for this snapshot.
