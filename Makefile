CC = gcc
CFLAGS = -Wall -Wextra -g

# Directories
ASM_DIR = src/assembler
EMU_DIR = src/emulator

# Source files
ASM_SRCS = $(ASM_DIR)/asm.c $(ASM_DIR)/parser.c $(ASM_DIR)/pass1.c $(ASM_DIR)/pass2.c $(ASM_DIR)/utils.c
EMU_SRCS = $(EMU_DIR)/emu.c $(EMU_DIR)/cpu.c $(EMU_DIR)/memory.c $(EMU_DIR)/trace.c

# Executables
ASM_EXE = asm
EMU_EXE = emu

all: $(ASM_EXE) $(EMU_EXE)

$(ASM_EXE): $(ASM_SRCS)
	$(CC) $(CFLAGS) -o $(ASM_EXE) $(ASM_SRCS)

$(EMU_EXE): $(EMU_SRCS)
	$(CC) $(CFLAGS) -o $(EMU_EXE) $(EMU_SRCS)

clean:
	rm -f $(ASM_EXE) $(EMU_EXE)
	rm -rf listings/ logs/ outputs/

.PHONY: all clean
