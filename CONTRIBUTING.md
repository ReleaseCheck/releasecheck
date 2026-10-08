# Contributing to ReleaseCheck

ReleaseCheck is being built phase by phase. Start with [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md), [ROADMAP.md](ROADMAP.md), and [docs/DESIGN.md](docs/DESIGN.md). Repository state is authoritative over stale chat context.

## Before changing code

1. Inspect the current phase and its acceptance criteria.
2. Check `git status --short --branch` and the repository tree.
3. Keep the change within the active phase unless the roadmap is updated deliberately.
4. Add deterministic offline tests for behavior and keep live registry tests separate.

## Development expectations

Use Go formatting, small interfaces, standard-library capabilities where practical, and explicit error handling. Never execute package code, install packages, invoke arbitrary builds, or use shell commands derived from package metadata. Do not add claims about adoption, novelty, security guarantees, or ecosystem validation without current evidence.

Run before submitting:

```text
gofmt -w .
go test ./...
go vet ./...
git diff --check
```

Use `go test -race ./...` when the change touches concurrency or shared state.

The default test suite is offline. Do not add tests that contact npm, PyPI, GitHub, or another public service unless they are explicitly opt-in and supplemental. Use checked-in inert fixtures or local test servers for deterministic behavior.

## Documentation and checkpoints

Important decisions must record what was chosen, why, alternatives considered, security implications, and future impact. At phase checkpoints update ROADMAP.md and PROJECT_CONTEXT.md, inspect the diff and tree, record limitations, and summarize verification. Do not commit or push automatically.

## Agent guidance

Antigravity with Claude Sonnet 4.6 is the default implementation agent. Gemini 3.1 Pro may help with broad planning. Copilot is secondary support. Use the strongest available Codex for independent security review, archive parsing, provenance ambiguity, difficult refactors, and the final audit. Do not default to OpenCode.

## Pull requests

Describe the phase, behavior, tests, security impact, documentation changes, and known limitations. Keep pull requests focused. A reviewer must be able to reproduce the result from the repository without chat history.

## Registry adapter work

Read [docs/REGISTRY_ADAPTERS.md](docs/REGISTRY_ADAPTERS.md) before proposing a new ecosystem. New registry support requires a scoped roadmap change, normalized evidence, offline fixtures, safe acquisition tests, documentation, and an explicit security review. Future ecosystems are not part of v0.1.
