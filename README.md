# SpryzeX

A custom assembly toolchain: **C two-pass assembler**, **C emulator**, and a **Go TUI IDE** for the SpryzeX ISA.

---

## Table of contents

- [Quick start](#quick-start)
- [Components](#components)
- [Assembler](#assembler)
- [Emulator](#emulator)
- [Spryzex IDE](#spryzex-ide)
- [Project structure](#project-structure)
- [Tests](#tests)
- [License](#license)

---

## Quick start

```bash
make
./asm samples/code.asm
./emu outputs/code.o
```

Or run the full pipeline and open a file in the IDE:

```bash
make
./spryzex-ide/spryzex-ide samples/code.asm
```

**Clean build artifacts:** `make clean` (removes binaries and `outputs/`, `logs/`, `listings/`)

---

## Components

| Component    | Language | Description |
|-------------|----------|-------------|
| **asm**     | C        | Two-pass assembler: `.asm` → `.o` |
| **emu**     | C        | Emulator: loads `.o`, runs instruction-by-instruction |
| **Spryzex IDE** | Go  | TUI: edit, assemble, run; calls `asm` and `emu` |

---

## Assembler

**Path:** `assembler/` · **Entry:** `asm.c`

### Usage

```bash
./asm program.asm
```

Creates `outputs/`, `logs/`, and `listings/` if needed. Produces:

- **outputs/\<base\>.o** — object file (only when there are no errors)
- **logs/\<base\>.log** — log messages
- **listings/\<base\>.lst** — listing with addresses

### Pipeline

1. **Read** — `read_source()` loads the file into a line buffer.
2. **Parse** — `parse_line()` strips `;` comments and fills `ParsedLine` (label, mnemonic, operand, line number).
3. **Pass 1** — Build symbol table, validate mnemonics and operand counts, record label references. Then check undefined/unused labels (undefined → error, unused → warning).
4. **Pass 2** — Encode instructions, resolve operands (including PC-relative branches). Write object file only if no errors; listing and log are always written.

**Encoding:** 8-bit opcode (low) + 24-bit operand (high):  
`(operand & 0xFFFFFF) << 8 | (opcode & 0xFF)`

### Instruction set

| Mnemonic | Opcode | Operand |
|----------|--------|---------|
| ldc, adc, ldl, stl, ldnl, stnl | 0–5 | immediate/offset |
| add, sub, shl, shr, a2sp, sp2a, return, HALT, out, outc | 6–9, 11–12, 14, 18, 21–22 | — |
| adj, call, brz, brlz, br | 10, 13, 15–17 | 1 or 2 (branch = PC-relative) |
| data, SET | 19–20 | data = raw word; SET = directive |

### Errors and warnings

- **Errors** (object file not written): invalid or duplicate label, undefined label, invalid mnemonic, missing/extra operand. Messages go through `add_log_entry()`; listing and log are still produced.
- **Warnings:** unused label (assembly can still succeed).
- **File open failure:** message to stdout, exit with 1.

---

## Emulator

**Path:** `emulator/` · **Entry:** `emu.c`

### Usage

```bash
./emu [flags] program.o
```

| Flag | Effect |
|------|--------|
| `-trace` | Instruction trace and register state |
| `-read` | Trace memory reads |
| `-write` | Trace memory writes |
| `-before` | Memory dump before execution |
| `-after` | Memory dump after execution |
| `-help` | Show help |

### Execution

1. **Init** — A=0, B=0, PC=0, SP=10000.
2. **Load** — Read 4-byte words from the binary into `memory[]` (size 2²⁴). Open failure → exit 1.
3. **Run** — Fetch-decode-execute loop; stops on **HALT** (opcode 18), optional execution limit, or after 2²⁴ steps (infinite-loop guard).

**CPU:** A, B, PC, SP. **Memory:** global array; stack at SP (e.g. ldl/stl use `memory[SP + operand]`).

### Errors

- Unknown opcode → message and exit 1.
- Object file open failure → message and return 1.
- Object too large → warning and stop loading.
- Infinite loop (2²⁴ instructions) → message and exit 1.

---

## Spryzex IDE

**Path:** `spryzex-ide/` · **Stack:** Go, [BubbleTea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss)

Modal, nvim-style TUI with a 3D rotating Spryzex mascot and orbiting Phobos (blue) and Deimos (gold) moons.

### Features

**Editor**

- Modes: **NORMAL** / **INSERT** / **VISUAL** / **COMMAND**
- Syntax highlighting (mnemonics, labels, numbers, comments, directives)
- Undo/redo (`u` / `Ctrl+R`), search (`/`, `n`/`N`), word motions (`w`, `b`, `e`)
- Auto-indent, inline diagnostics from assembler errors

**Console**

- **LIVE** — real-time assembler/emulator output  
- **LOG** — assembler log file  
- **LISTING** — `.lst` with addresses  
- **OBJ** — disassembled object (hex)

**Mascot**

- 3D sphere (Andy Sloane–style). States: **IDLE**, **BUILDING**, **SUCCESS**, **ERROR**, **RUNNING**

**Preprocessor (Go layer, C untouched)**

- `.equ NAME = VALUE`, `.macro` / `.endmacro`, `.include "file.asm"`, constant substitution

### Key bindings

**Normal**

| Key | Action |
|-----|--------|
| `i` / `I` / `a` / `A` | Insert (at cursor / line start / after / line end) |
| `o` / `O` | New line below / above |
| `v` / `:` | Visual mode / Command mode |
| `h j k l` | Move |
| `w` / `b` / `e` | Word forward / back / end |
| `0` / `$` / `^` | Line start / end / first non-space |
| `gg` / `G` | Top / bottom of file |
| `x` / `dd` / `D` / `C` | Delete char / line / to EOL / change to EOL |
| `u` / `Ctrl+R` | Undo / Redo |
| `/` | Search |
| `B` / `R` / `S` | Build / Run / Save |

**Command (`:`)** — `:w` `:q` `:wq` `:q!` `:w path` `:e path` `:build` `:run`

**Global** — `Ctrl+B` Build · `Ctrl+R` Run · `Ctrl+S` Save · `Ctrl+W` Focus panel · `Ctrl+]` / `Ctrl+[` Tabs · `Ctrl+C` Quit

### Install and run

From repo root:

```bash
make
./spryzex-ide/spryzex-ide [path/to/file.asm]
```

Or from `spryzex-ide/`:

```bash
bash install.sh
# or: go build -o spryzex-ide . && ./spryzex-ide your_file.asm
```

Requires **Go 1.22+**. The IDE looks for `asm` and `emu` under the project root or in `/usr/local/bin`; if missing, it can run `make` to build them.

### Theme (palette)

| Name | Hex | Use |
|------|-----|-----|
| SpryzexRed | `#C1440E` | Normal mode, sphere dark |
| SpryzexBright | `#E85D26` | Cursor line, borders |
| SpryzexGlow | `#FF6B35` | Cursor, logo |
| PhobosBlue | `#4A9EFF` | Insert, labels |
| DeimosGold | `#FFD700` | Numbers, Deimos |
| NebulaPurp | `#B48EAD` | Directives, trace |
| AuroraGreen | `#A3BE8C` | Success, console |
| CometCyan | `#88C0D0` | Registers |

---

## Project structure

```
SpryzeX/
├── README.md
├── Makefile
├── LICENSE
├── assembler/              # C assembler
│   ├── asm.c, asm.h, state.h
│   ├── parser.c/h, pass1.c/h, pass2.c/h
│   └── utils.c/h
├── emulator/               # C emulator
│   ├── emu.c
│   ├── cpu.c/h, memory.c, trace.c
├── spryzex-ide/            # Go TUI
│   ├── main.go
│   ├── internal/
│   │   ├── theme/
│   │   ├── editor/
│   │   ├── spryzex/
│   │   └── assembler/      # bridge + preprocessor
│   └── install.sh
├── samples/
│   └── code.asm
└── tests/                  # Assembly test sources
```

Generated (not committed): `asm`, `emu`, `outputs/`, `logs/`, `listings/`, `spryzex-ide/spryzex-ide`.

---

## Tests

Assembly sources live in `tests/`. Run by hand:

```bash
./asm tests/<name>.asm
./emu outputs/<name>.o
```

Examples: `test_add.asm`, `test_undefined_label.asm`, `test_duplicate_label.asm`, `bubble_sort.asm`, `factorial_recursion.asm`, etc.

---

## License

See [LICENSE](LICENSE).
