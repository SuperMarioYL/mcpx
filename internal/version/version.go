// Package version holds the single authoritative product version string.
package version

// Version is the product version. Every user-visible version surface — the
// VERSION file, `mcpx --version`, the CHANGELOG head, web/site.json's
// content_version and source_version — and the MCP clientInfo field sent by the
// handshake must agree with it; TestVersionLockstep enforces that. Bump this
// one constant and update VERSION/CHANGELOG/site.json with it.
const Version = "0.7.0"
