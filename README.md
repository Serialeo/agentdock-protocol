# agentdock-protocol

Shared contracts used by AgentDock and NexusDock.

The repository intentionally has two logical layers:

- the root `protocol` package owns only the AgentDock ↔ NexusDock Bridge wire protocol: envelopes, Hello capabilities, operations, and MCP App resource identities/contracts;
- `mcpcontract` owns only the canonical model-facing MCP contracts shared by the two entrypoints: input/output schemas, annotations, and bounded behavior vectors.

Neither package owns AgentDock runtime behavior, NexusDock stores, renderer HTML, Recall persistence, Workflow persistence, or HTTP handlers. Those remain in their application repositories.

## Releases

Every push and pull request runs formatting, module tidiness, vet, tests, race tests, and compilation. Push a version tag such as `v1.2.3` to rerun these checks on the tagged commit and publish a GitHub Release with generated notes and GitHub-provided source archives. Tags may include a SemVer prerelease suffix, such as `v1.2.3-rc.1`; these releases are marked as prereleases and never marked as latest. Build metadata in tags is not supported. This shared Go library does not publish a Docker image.

Publish the protocol version tag first, then explicitly update and verify the `agentdock-protocol` dependency in AgentDock and NexusDock before releasing those applications. This workflow does not create or move version tags and does not automatically upgrade either application repository. Publishing uses the repository's built-in `GITHUB_TOKEN` with `contents: write`; no separate release secret is required.

## WorkSession continuation contracts

Continuation is a Nexus-owned extension of the existing Project / WorkSession / Target flow. Bridge remains version 4: nodes opt into `bridge.command.outcomes.v1` to support bounded `command.outcomes.read` and idempotent `command.outcomes.ack`. Command outcomes retain their execution context, stable event identity, durable output snapshot, and pending-report marker. Explicit session reads also return already acknowledged outcomes, covering commands that finish before `await`. Terminal persistence does not guarantee that a process survives a node restart, and an unknown outcome never requests an automatic rerun.

The three model entrances (`work_continuation`, `present_work_continuation`, `consume_work_wake`) have explicit model visibility; controller coordination tools are app-only. Only `present_work_continuation` binds `ui://agentdock/work-continuation` (`agentdock.work-continuation.v1`). The controller's binding proof belongs exclusively in tool result `_meta["agentdock/work-continuation"]`, outside model-visible content. Resume messages contain only the exact seven-field identity envelope and a single-use consume token; actual work and routing remain server-authorized. Explicit enable/recovery, bounded rounds, prepared-message fencing, and consumed-versus-settled state are separate contracts.

`go test ./...` covers wire/schema compatibility; `node --test mcpapps/continuation_test.mjs` exercises the controller against a simulated public MCP Apps Host transport. CI runs both suites. Live ChatGPT scheduling and lifecycle behavior require separate Host acceptance testing.

Built-in capability snapshots use `Hello.builtins` and the complete `node.updated`
message. `provided`, `enabled`, and `ready` are independent; a transitioning group
is unavailable. The node owns and persists the user choice. A snapshot replaces
the tool catalog; it never grants Deployment permissions or resumes old work.
GUI updates use the existing `runtime.request` operation for
`/internal/runtime/builtins`, with `{ "id": "browser", "enabled": false }`.
