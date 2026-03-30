# Implementation Tasks: HTML Flamegraph Export

## Tasks

- [x] 1. Wire profiling flag into SimulationRequest in debug.go
  - In `runDebug`, set `Profile: ProfileFlag` on the `SimulationRequest` when building it
  - **Acceptance**: Requirements 1.1, 1.2, 1.3

- [x] 2. Add flamegraph export helpers to debug.go
  - Add `resolveExportFormat(flag string) visualizer.ExportFormat` function
  - Add `exportFlamegraphIfNeeded(txHash string, resp *simulator.SimulationResponse) error` function
  - Call `exportFlamegraphIfNeeded` after simulation completes in `runDebug`
  - **Acceptance**: Requirements 2.1–2.5, 3.1–3.4, 6.1–6.2

- [x] 3. Unit tests for debug.go helpers
  - `TestResolveExportFormat` — html → FormatHTML, svg → FormatSVG, unknown → FormatHTML + warning
  - `TestExportFlamegraphIfNeeded_NoProfile` — no file written when ProfileFlag false
  - `TestExportFlamegraphIfNeeded_EmptyFlamegraph` — warning printed, no file written
  - `TestExportFlamegraphIfNeeded_WritesFile` — file created with correct name and content

- [x] 4. Property-based tests for visualizer
  - Property 1: `GenerateInteractiveHTML` output is self-contained (no external HTTP refs, contains DOCTYPE and SVG)
  - Property 2: `ExportFlamegraph` preserves SVG content for any format
  - Property 3: `InjectDarkMode` is idempotent
