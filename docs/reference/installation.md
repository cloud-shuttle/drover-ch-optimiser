---
title: "Installation"
description: "Install drover-ch with go install or from source."
product: drover-ch-optimiser
audience: platform-operator
doc_type: reference
topics:
  - data-warehousing
surface: repo-docs
---

# Installation

## Go install

```bash
go install github.com/cloud-shuttle/drover-ch-optimiser/cmd/drover-ch@latest
```

## Release binaries

Tagged releases publish archives for Linux, macOS, and Windows (`amd64` and `arm64`).

1. Open [GitHub Releases](https://github.com/cloud-shuttle/drover-ch-optimiser/releases).
2. Download the archive for your OS/architecture.
3. Extract `drover-ch` onto your `PATH`.

Verify:

```bash
drover-ch --version
drover-ch rules
```

## Homebrew (optional)

A sample formula lives at `packaging/homebrew/drover-ch.rb`. After the first release tag, update the `url` and `sha256`, then:

```bash
brew install ./packaging/homebrew/drover-ch.rb
```

For a project tap, copy the formula into `homebrew-tap/Formula/drover-ch.rb` and point `url` at the release tarball.

## Documentation site

Published from `main` via GitHub Pages: enable **Pages → GitHub Actions** in the repository settings.

Local preview:

```bash
pip install mkdocs-material
mkdocs serve
```

Deploy manually (maintainers):

```bash
mkdocs gh-deploy --force
```
