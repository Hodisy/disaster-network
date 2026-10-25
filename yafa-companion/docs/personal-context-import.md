# Private personal-context import

## Implemented scope

`yafa-context-import` is an explicit maintenance executable. It reads the prepared private bundle, checks every manifest/catalog hash and size, maps local project/item references, and uses `WorkspaceCore.ImportPersonalContext` with the real memory repository. It never calls a provider and is not an AI tool or an API startup side effect. Private source files are not embedded in the image or committed to Git.

The import appends bounded marked sections to existing `user.md` and `directive.md`, preserving their original contents/revisions. Capacity or conflicts fail before writing. It creates/reuses matching active projects, adds scheduled human tasks with their notes, and retains event typing, original status, provenance and schedule precision in the import receipt. Reported completed tasks are checked; no completion instant, old conversation, run, notification or cron is fabricated. In-progress is retained in source metadata; the current checkbox schema has no separate in-progress field.

The documents and receipt are additive fields of the existing workspace aggregate. One expected-version compare-and-write commits all imported changes atomically, within the existing 196,608-byte aggregate cap. Unknown additive workspace fields are preserved through the maintenance round-trip rather than discarded. An active run lease blocks the import. The importer does not stop another run, reset records or overwrite task edits. A repeated import ID with the same fingerprint returns already_applied without replaying any old dates, notes or completion states; a changed package with the same ID conflicts. Do not change the ID merely to defeat a conflict and duplicate the same events.

## Available context after import

The runtime injects only the bounded `current-context` document (the prepared now.md, at most 1,600 UTF-8 bytes), its as-of/stale state and a catalog lookup hint. Profile/preferences are included by the existing versioned context assembly. Journals and the original account are not dumped into every prompt. An as-of date different from today is explicitly stale, not silently rewritten as current state.

Two commands join the existing filtered command registry:

```text
memory.list --offset 0 --limit 10
memory.read --id <id-returned-by-list> --offset 0 --limit 4096
```

Listing returns metadata and section names only. Reads return at most 4,096 UTF-8 bytes of body, hash/revision/source and next_offset; offsets must remain at character boundaries. A caller follows the returned cursor instead of guessing. IDs resolve from stored owner-scoped documents, never filesystem paths or URLs. Retrieved notes are data, not tool grants or higher-priority instructions. Existing run/tool/input/token bounds remain in force; this slice does not add a separate cumulative 8 KiB retrieval quota, semantic search, section-name filter, memory-write tools or a task-reschedule command.

## Prepare, review, apply

Use the same new backend version for the API and importer. **Do not run an older workspace writer after importing**: older executables do not know the additive memory/receipt fields and may discard them on their next aggregate save. This is a backend-code rollout prerequisite, not a database reset or automatic deployment.

The prepared input lives under the user's ignored `.private/context/import-2026-10-25/` folder in the original checkout. It must be supplied privately at runtime. That folder is not supplied by GitHub or the image. The manifest, catalog and tasks schemas are documented by the bundle itself.

Offline integrity inspection (no credential, database or provider call):

```sh
yafa-context-import --bundle /private/import-bundle --verify-only
```

This checks package integrity/shape, not whether a particular target can merge it. A read-only plan uses the private `MEMORY_URL` and `YAFA_MEMORY_TOKEN` already provided by the execution environment. Never put the token in command arguments, logs, Git or a chat.

```sh
yafa-context-import --bundle /private/import-bundle
```

Review the counts, workspace_version and fingerprint. If the target or package changes, plan again. Only an explicit apply writes:

```sh
yafa-context-import --bundle /private/import-bundle \
  --apply --expected-version '<reviewed-workspace-version>' \
  --expected-fingerprint '<reviewed-bundle-fingerprint>'
```

A stale version fails; there is no automatic refresh-and-apply. After a storage/transport ambiguity, rerun a read-only plan against the same bundle to discover its actual receipt. Readback errors after a commit are reported as committed but unconfirmed; do not invent rollback. Credential values and personal bodies are not printed by this command. A successful applied=true means the compare-and-write and receipt readback succeeded, not that a phone/widget has refreshed.

The Dockerfile packages the maintenance executable alongside the API, without changing the normal entrypoint. From the existing infra directory, an owner can run a one-off command with a private read-only mount and the same Compose environment/network:

```sh
docker compose --env-file .env -f compose.yaml run --rm --no-deps \
  -v '/absolute/private/import-folder:/private/import-bundle:ro' \
  --entrypoint yafa-context-import api --bundle /private/import-bundle
```

Append the explicit apply/version/fingerprint flags only after reviewing the plan. These examples are not commands executed by the coding agent. No actual personal import, app launch, service/container restart, paid call or database write is part of source preparation.
