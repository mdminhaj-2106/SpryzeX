CC = gcc
CFLAGS = -Wall -Wextra -g

# Directories
ASM_DIR = src/assembler
EMU_DIR = src/emulator
TUI_DIR = tui

# Source files
ASM_SRCS = $(ASM_DIR)/asm.c $(ASM_DIR)/parser.c $(ASM_DIR)/pass1.c $(ASM_DIR)/pass2.c $(ASM_DIR)/utils.c
EMU_SRCS = $(EMU_DIR)/emu.c $(EMU_DIR)/cpu.c $(EMU_DIR)/memory.c $(EMU_DIR)/trace.c
TUI_SRCS = $(TUI_DIR)/main.c $(TUI_DIR)/ui.c $(TUI_DIR)/editor.c $(TUI_DIR)/runner.c $(TUI_DIR)/mascot.c $(TUI_DIR)/state.c

# Executables
ASM_EXE = asm
EMU_EXE = emu
TUI_EXE = spryzex

NCURSES_LIBS ?= -lncurses -lpanel -lmenu

all: $(ASM_EXE) $(EMU_EXE) $(TUI_EXE)

$(ASM_EXE): $(ASM_SRCS)
	$(CC) $(CFLAGS) -o $(ASM_EXE) $(ASM_SRCS)

$(EMU_EXE): $(EMU_SRCS)
	$(CC) $(CFLAGS) -o $(EMU_EXE) $(EMU_SRCS)

$(TUI_EXE): $(TUI_SRCS)
	$(CC) $(CFLAGS) -o $(TUI_EXE) $(TUI_SRCS) $(NCURSES_LIBS)

clean:
	rm -f $(ASM_EXE) $(EMU_EXE) $(TUI_EXE)
	rm -rf listings/ logs/ outputs/

.PHONY: all clean
