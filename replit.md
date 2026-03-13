# SpryzeX Toolchain + Spryzex IDE

## Overview

SpryzeX is a custom assembly language toolchain with three components:

1. **C Assembler (`asm`)** — A two-pass assembler for the SpryzeX ISA
2. **C Emulator (`emu`)** — A CPU emulator that executes assembled object files
3. **Spryzex IDE (`spryzex-ide/spryzex-ide`)** — A Go BubbleTea TUI IDE for editing, assembling, and running assembly programs

## Architecture

- **Language**: C (assembler/emulator) + Go 1.21 (IDE)
- **Build System**: GNU Make
- **Go Dependencies**: BubbleTea (TUI), Lipgloss (styling)

## Project Structure

```
assembler/       - C two-pass assembler source
emulator/        - C emulator source
spryzex-ide/     - Go BubbleTea TUI IDE
  main.go        - Main app, layout, key handling
  go.mod         - Go module file
  go.sum         - Go dependency checksums
  internal/
    theme/       - Color palette + lipgloss styles
    editor/      - Modal editor (nvim-style)
    spryzex/     - 3D planet mascot renderer
    assembler/   - Bridge to C assembler/emulator + preprocessor
samples/         - Sample assembly files
tests/           - Test assembly programs
outputs/         - Generated object files (gitignored)
listings/        - Assembler listing files (gitignored)
logs/            - Assembler log files (gitignored)
```

## Build

```bash
make          # builds asm, emu, and spryzex-ide/spryzex-ide
make clean    # remove binaries and generated output
```

## Run

```bash
./asm samples/code.asm        # Assemble a file
./emu outputs/code.o          # Run the emulator
./spryzex-ide/spryzex-ide     # Launch the TUI IDE
```

## Workflow

The project runs as a console workflow that builds and launches the TUI IDE:

```
make && ./spryzex-ide/spryzex-ide samples/code.asm
```

## Notes

- The `go.mod` file was created during Replit import (was missing from the original repo)
- The `github.com/golang/sync` and `github.com/golang/sys` entries in go.sum are aliases; actual imports use `golang.org/x/sync` and `golang.org/x/sys`
- This is a TUI app — no web frontend, no port binding
