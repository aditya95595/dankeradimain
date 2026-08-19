---
name: Go publish toolchain compatibility
description: Publishing uses the preinstalled Go toolchain and cannot rely on automatic toolchain downloads.
---

The production build must target the Go version available in the publish environment rather than relying on Go's automatic toolchain download, because the publish network may not reach the checksum/toolchain service.

**Why:** A newer `go` directive caused publishing to enter a crash loop before the application binary started.

**How to apply:** Keep `go.mod` compatible with the installed publish toolchain and verify the actual production build command locally before publishing.