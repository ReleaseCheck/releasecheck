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

The corpus intentionally includes both a strong npm source reference and PyPI distributions where the source repository is visible but an exact release commit is not supplied. That distinction is a product behavior, not a test failure.
