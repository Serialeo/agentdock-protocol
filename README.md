# agentdock-protocol

Shared contracts used by AgentDock and NexusDock.

The repository intentionally has two logical layers:

- the root `protocol` package owns only the AgentDock ↔ NexusDock Bridge wire protocol: envelopes, Hello capabilities, operations, and MCP App resource identities/contracts;
- `mcpcontract` owns only the canonical model-facing MCP contracts shared by the two entrypoints: input/output schemas, annotations, and bounded behavior vectors.

Neither package owns AgentDock runtime behavior, NexusDock stores, renderer HTML, Recall persistence, Workflow persistence, or HTTP handlers. Those remain in their application repositories.

## Releases

Every push and pull request runs formatting, module tidiness, vet, tests, race tests, and compilation. Push a version tag such as `v1.2.3` to rerun these checks on the tagged commit and publish a GitHub Release with generated notes and GitHub-provided source archives. Tags may include a SemVer prerelease suffix, such as `v1.2.3-rc.1`; these releases are marked as prereleases and never marked as latest. Build metadata in tags is not supported. This shared Go library does not publish a Docker image.

Publish the protocol version tag first, then explicitly update and verify the `agentdock-protocol` dependency in AgentDock and NexusDock before releasing those applications. This workflow does not create or move version tags and does not automatically upgrade either application repository. Publishing uses the repository's built-in `GITHUB_TOKEN` with `contents: write`; no separate release secret is required.
