# SpryzeX Toolchain + Spryzex IDE

SpryzeX provides:
- A C two-pass assembler (`asm`)
- A C emulator (`emu`)
- A Go BubbleTea TUI in [`spryzex-ide/`](./spryzex-ide)

## Build

```bash
make
```

This builds:
- `asm`
- `emu`
- `spryzex-ide/spryzex-ide`

## Run

Assemble:
```bash
./asm samples/code.asm
```

Emulate:
```bash
./emu outputs/code.o
```

Launch IDE:
```bash
./spryzex-ide/spryzex-ide [optional/path/to/file.asm]
```

## Spryzex IDE

The TUI is organized as:

```text
spryzex-ide/
├── main.go
├── internal/
│   ├── theme/theme.go
│   ├── editor/editor.go
│   ├── spryzex/spryzex.go
│   └── assembler/
│       ├── assembler.go
│       └── preprocessor.go
└── install.sh
```

For full IDE feature details and keybindings, see [`spryzex-ide/README.md`](./spryzex-ide/README.md).

## Clean

```bash
make clean
```

This removes binaries and generated output folders (`outputs/`, `logs/`, `listings/`).
