/*
 * emu.c - SpryzeX emulator CLI (load object, run, trace, memory dump)
 * Author: [YOUR FULL NAME]
 * User ID: [YOUR USER ID]
 * Declaration: I declare that this code is my own work.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "cpu.h"

void display_help() {
    printf("SpryzeX Emulator\n");
    printf("Usage: ./emu [flags] program.o\n");
    printf("Flags (can be combined):\n");
    printf("  -trace    Show instruction trace with register states\n");
    printf("  -read     Show memory read operations\n");
    printf("  -write    Show memory write operations\n");
    printf("  -before   Show memory dump before execution\n");
    printf("  -after    Show memory dump after execution\n");
    printf("  -help     Show this help message\n");
}

int main(int argc, char *argv[])
{
    if(argc < 2)
    {
        display_help();
        return 1;
    }

    char *filename = NULL;
    int trace_mode = 0;   /* 0: none, 1: trace, 2: read, 3: write */
    int before_dump = 0;
    int after_dump = 0;

    int i;
    for (i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-help") == 0) {
            display_help();
            return 0;
        }
        if (strcmp(argv[i], "-trace") == 0)  trace_mode = 1;
        else if (strcmp(argv[i], "-read") == 0)   trace_mode = 2;
        else if (strcmp(argv[i], "-write") == 0) trace_mode = 3;
        else if (strcmp(argv[i], "-before") == 0) before_dump = 1;
        else if (strcmp(argv[i], "-after") == 0)  after_dump = 1;
        else if (argv[i][0] == '-') {
            printf("Unknown flag: %s\n", argv[i]);
            display_help();
            return 1;
        } else {
            filename = argv[i];
        }
    }

    if (filename == NULL) {
        printf("Error: No program file specified\n");
        display_help();
        return 1;
    }

    init_cpu();

    int size = load_object(filename);
    if (size < 0) return 1;

    printf("Loaded %d instructions into memory\n", size);

    if (before_dump) {
        printf("\n--- Memory dump BEFORE execution ---\n");
        dump_memory(size + 10);
        printf("-------------------------------------\n\n");
    }

    run_cpu(0, trace_mode);

    if (after_dump) {
        printf("\n--- Memory dump AFTER execution ---\n");
        dump_memory(size + 10);
        printf("------------------------------------\n");
    }

    return 0;
}
