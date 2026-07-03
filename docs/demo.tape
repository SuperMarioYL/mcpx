# vhs script — happy path for mcpx.
# Render locally:  vhs docs/demo.tape   (writes assets/demo.gif)
# Or use the `demo` GitHub workflow (workflow_dispatch).

Output assets/demo.gif

Set Shell    "bash"
Set FontSize 18
Set Width    1100
Set Height   680
Set Padding  24
Set Theme    "Catppuccin Mocha"
Set TypingSpeed 55ms

# Assumes `mcpx` is on PATH (go install ./cmd/mcpx) and at least one client is detected.

Type "mcpx catalog"
Enter
Sleep 2.5s

Type "mcpx add filesystem --yes"
Enter
Sleep 13s

Type "mcpx list"
Enter
Sleep 4s
