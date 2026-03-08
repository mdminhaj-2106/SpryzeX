# SpryzeX: Modular Two-Pass Assembler & Emulator

Welcome to **SpryzeX**, a minimal yet powerful toolchain for a custom MIPS-like Instruction Set Architecture (ISA).

---

## 🚀 Quick Start

### 1. Build the System
Use the provided `Makefile` to compile both the assembler and the emulator:
```bash
make
```
This generates two executables: `asm` (Assembler) and `emu` (Emulator).

### 2. Assemble a Program
Convert your assembly (`.asm`) source into machine code (`.o`):
```bash
./asm src/tests/bubble_sort.asm
```
- **Outputs:** `outputs/` folder (binary), `logs/` folder (errors/warnings), `listings/` folder (human-readable encoding).

### 3. Run the Emulator
Execute the generated machine code:
```bash
./emu outputs/bubble_sort.o
```

**Emulator Flags:**
- `-trace`: See every instruction executed and the state of registers (PC, SP, A, B).
- `-read`: Track memory read operations.
- `-write`: Track memory write operations.
- `-before`: Dump memory state *before* execution.
- `-after`: Dump memory state *after* execution.

---

## 🏗️ System Architecture

### 1. The Assembler (`src/assembler/`)
A **Two-Pass Assembler** that ensures all label references (even forward ones) are resolved correctly.
- **Pass 1:** Scans the code to build a **Symbol Table**, calculating addresses for every label.
- **Pass 2:** Generates the final machine code by looking up opcodes and resolving label addresses into offsets.
- **Directives:** Supports `data` (storing constants) and `SET` (defining constants).

### 2. The Emulator (`src/emulator/`)
A modular CPU simulation with:
- **Registers:** 
  - `A`: Accumulator (Primary register).
  - `B`: Secondary Accumulator (used for math and temporary storage).
  - `PC`: Program Counter.
  - `SP`: Stack Pointer (initialized to 10,000 to keep stack far from code).
- **Memory:** 16MB addressable space.
- **Components:** Modularized into `cpu.c` (logic), `memory.c` (storage), and `trace.c` (debugging).

---

## 📜 Instruction Set (ISA)

| Opcode | Mnemonic | Type | Description |
|--------|----------|------|-------------|
| 0      | `ldc`    | 1    | Load constant into A, B = old A |
| 1      | `adc`    | 1    | Add constant to A |
| 2      | `ldl`    | 1    | Load from Stack: A = mem[SP + offset], B = old A |
| 3      | `stl`    | 1    | Store to Stack: mem[SP + offset] = A, A = B |
| 4      | `ldnl`   | 1    | Load Non-Local: A = mem[A + offset] |
| 5      | `stnl`   | 1    | Store Non-Local: mem[A + offset] = B |
| 6      | `add`    | 0    | A = B + A |
| 7      | `sub`    | 0    | A = B - A |
| 8      | `shl`    | 0    | A = B << A |
| 9      | `shr`    | 0    | A = B >> A |
| 10     | `adj`    | 1    | Adjust Stack Pointer: SP = SP + offset |
| 13     | `call`   | 2    | Call function: B = A, A = PC, PC = PC + offset |
| 14     | `return` | 0    | Return: PC = A, A = B |
| 15     | `brz`    | 2    | Branch if A == 0 |
| 16     | `brlz`   | 2    | Branch if A < 0 |
| 17     | `br`     | 2    | Unconditional Branch |
| 18     | `HALT`   | 0    | Stop execution |
| 21     | `out`    | 0    | Print A as signed integer (newline) |
| 22     | `outc`   | 0    | Print low 8 bits of A as a character |

---

## 📁 Folder Structure
- `src/assembler/`: Source for the assembler.
- `src/emulator/`: Source for the emulator.
- `src/tests/`: Edge case tests (Sorting, Primes, Arithmetic, Loops).
- `Formats&Flows/`: Detailed Mermaid flowcharts explaining the internals.
- `outputs/`, `logs/`, `listings/`: Automatically generated during assembly.

---

## 🛠️ Maintenance
To clean up all build artifacts and logs:
```bash
make clean
```
