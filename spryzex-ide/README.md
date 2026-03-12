# ◈ SPRYZEX IDE

A **modal, nvim-style assembly language IDE** built in Go with BubbleTea + Lipgloss.  
Features a **3D rotating Spryzex** mascot with orbiting Phobos and Deimos moons.

---

## Stack

| Layer      | Technology                                |
|------------|-------------------------------------------|
| TUI        | Go + [BubbleTea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Assembler  | Your existing **C assembler** (untouched) |
| Emulator   | Your existing **C emulator** (untouched)  |
| Theme      | Custom Spryzex/Deep Space palette            |

---

## Features

### Editor (nvim-style modal)
- `NORMAL` / `INSERT` / `VISUAL` / `COMMAND` modes
- Full syntax highlighting (mnemonics, labels, numbers, comments, directives)
- Multi-level undo/redo (`u` / `Ctrl+R`)
- `/` search with `n`/`N` navigation
- Word motions (`w`, `b`, `e`)
- Auto-indent on newline
- Inline diagnostics from assembler errors

### Spryzex Mascot (3D ASCII)
- Real 3D sphere rendered with Andy Sloane-style math
- **Phobos** (blue) and **Deimos** (gold) orbit in ellipses
- Animation states:
  - 🔴 **IDLE** — calm gentle spin
  - 🔥 **BUILDING** — fast spin with fire particles shooting out
  - ✅ **SUCCESS** — star burst explosion
  - ❌ **ERROR** — red pulse with error markers
  - ▶️  **RUNNING** — smooth emulation orbit

### Console
Four output tabs:
- **LIVE** — real-time assembler/emulator output  
- **LOG** — assembler log file  
- **LISTING** — `.lst` listing file with addresses  
- **OBJ** — disassembled object file with hex

### Enhanced Assembler (pre-processor layer)
Zero changes to your C code. The Go layer adds:
- `.equ NAME = VALUE` — named constants
- `.macro NAME param1 ... / .endmacro` — macro definitions
- `.include "file.asm"` — file inclusion
- Constant substitution throughout source

---

## Key Bindings

### Normal Mode
| Key | Action |
|-----|--------|
| `i` | Insert mode |
| `I` | Insert at line start |
| `a` | Insert after cursor |
| `A` | Insert at line end |
| `o` / `O` | New line below / above |
| `v` | Visual mode |
| `:` | Command mode |
| `h j k l` | Move cursor |
| `w` / `b` / `e` | Word forward / back / end |
| `0` / `$` / `^` | Line start / end / first non-space |
| `gg` / `G` | File top / bottom |
| `Ctrl+D` / `Ctrl+U` | Half-page down / up |
| `x` | Delete char |
| `d` | Delete line (dd) |
| `D` | Delete to end of line |
| `C` | Change to end of line |
| `u` | Undo |
| `Ctrl+R` | Redo |
| `/` | Search |
| `B` | Build (assemble) |
| `R` | Run (emulator) |
| `S` | Save |

### Command Mode (`:`)
| Command | Action |
|---------|--------|
| `:w` | Save |
| `:q` | Quit |
| `:wq` / `:x` | Save and quit |
| `:q!` | Force quit |
| `:w path` | Save as |
| `:e path` | Open file |
| `:build` | Assemble |
| `:run` | Run emulator |

### Global (any mode)
| Key | Action |
|-----|--------|
| `Ctrl+B` | Build |
| `Ctrl+R` | Run |
| `Ctrl+S` | Save |
| `Ctrl+W` | Cycle panel focus |
| `Ctrl+]` | Next output tab |
| `Ctrl+[` | Previous output tab |
| `Ctrl+C` | Quit |

---

## Install

```bash
# From the spryzex-ide directory:
bash install.sh

# Or manually:
cd spryzex-ide
go build -o spryzex-ide .
./spryzex-ide your_file.asm
```

**Requires Go 1.22+**

---

## Project Structure

```
spryzex-ide/
├── main.go                     # BubbleTea app, layout, key handling
├── internal/
│   ├── theme/theme.go          # Spryzex color palette + lipgloss styles
│   ├── editor/editor.go        # Modal editor, syntax highlight, undo
│   ├── spryzex/spryzex.go            # 3D planet renderer + animations
│   └── assembler/
│       ├── assembler.go        # Bridge to C assembler/emulator
│       └── preprocessor.go     # Macro/include/equ expansion
└── install.sh                  # Build + install script
```

---

## Color Palette

| Name | Hex | Use |
|------|-----|-----|
| SpryzexRed | `#C1440E` | Normal mode, sphere dark |
| SpryzexBright | `#E85D26` | Cursor line, active borders |
| SpryzexGlow | `#FF6B35` | Cursor, logo, notifications |
| PhobosBlue | `#4A9EFF` | Insert mode, labels, info |
| DeimosGold | `#FFD700` | Numbers, Deimos moon |
| NebulaPurp | `#B48EAD` | Directives, trace output |
| AuroraGreen | `#A3BE8C` | Success, strings, console border |
| CometCyan | `#88C0D0` | Registers, secondary |
