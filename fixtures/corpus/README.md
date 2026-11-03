# Real Package Corpus

This directory records pinned public npm/PyPI examples for reproducible compatibility checks. It stores metadata and expected digests, not installed packages or executable package code.

The records are supplemental to the offline test suite. They are not fetched during default CI because public registries, source hosts, and release metadata can change or become unavailable. A live corpus check must be explicit, network-enabled, bounded, and non-executing.

Run the opt-in verifier from the repository root:

```text
go run ./cmd/corpuscheck -output-dir .tmp/corpus
```

It downloads only the pinned URLs, checks HTTPS, redirect limits, size bounds, expected sizes, and SHA-256 digests, then stores the verified bytes in the requested directory. It never installs, extracts, imports, or executes an artifact. `.tmp/` is ignored by Git and should be deleted when no longer needed.

Each record includes:

- registry, package, and version;
- selected artifact filename, URL, size, and SHA-256 or registry integrity;
- source repository and exact source reference when the registry publishes one;
- the expected evidence state and comparison limitation;
- the observation date and primary metadata URL.

The real corpus intentionally includes npm packages with and without exact source commits, plus PyPI sdists and wheels. The package selection is adversarial to the verifier's evidence model: it exercises generated distribution files, multiple release-file kinds, missing exact source references, and source-bound releases. It does not make a maliciousness claim about any package.

`adversarial-cases.json` is the offline companion matrix. It names the controlled fixtures for artifact-only, source-only, modified, unsafe-path, install-metadata, and provenance-subject-mismatch evidence. These cases are deterministic and run in default CI; the public release records are opt-in because registries and source hosts are external dependencies.
