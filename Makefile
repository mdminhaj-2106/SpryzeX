CC     = gcc
CFLAGS = -Wall -Wextra -g -std=c99
GO     = go

# ── Directories ───────────────────────────────────────────
ASM_DIR = assembler
EMU_DIR = emulator
SPRYZEX_DIR = spryzex-ide

# ── Source files ──────────────────────────────────────────
ASM_SRCS = $(ASM_DIR)/asm.c \
           $(ASM_DIR)/parser.c \
           $(ASM_DIR)/pass1.c \
           $(ASM_DIR)/pass2.c \
           $(ASM_DIR)/utils.c

EMU_SRCS = $(EMU_DIR)/emu.c \
           $(EMU_DIR)/cpu.c \
           $(EMU_DIR)/memory.c \
           $(EMU_DIR)/trace.c

# ── Executables ───────────────────────────────────────────
ASM_EXE = asm
EMU_EXE = emu
SPRYZEX_EXE = $(SPRYZEX_DIR)/spryzex-ide

# ═════════════════════════════════════════════════════════
all: $(ASM_EXE) $(EMU_EXE) spryzex-ide

$(ASM_EXE): $(ASM_SRCS)
	$(CC) $(CFLAGS) -o $@ $^

$(EMU_EXE): $(EMU_SRCS)
	$(CC) $(CFLAGS) -o $@ $^

spryzex-ide:
	cd $(SPRYZEX_DIR) && $(GO) build -o spryzex-ide .

clean:
	rm -f $(ASM_EXE) $(EMU_EXE) $(SPRYZEX_EXE)
	rm -rf listings/ logs/ outputs/
	rm -rf asm.dSYM emu.dSYM spryzex.dSYM

# Run the Hello-World sample through the full pipeline
demo: $(ASM_EXE) $(EMU_EXE)
	./asm samples/code.asm
	./emu outputs/code.o

.PHONY: all clean demo spryzex-ide
