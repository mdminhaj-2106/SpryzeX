#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "cpu.h"

void display_help() {
    printf("SpryzeX Emulator\n");
    printf("Usage: ./emu [flag] program.o\n");
    printf("Flags:\n");
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

    char *flag = NULL;
    char *filename = NULL;
    int trace_mode = 0; // 0: none, 1: trace, 2: read, 3: write

    if (argc == 2) {
        filename = argv[1];
    } else {
        flag = argv[1];
        filename = argv[2];

        if (strcmp(flag, "-trace") == 0) trace_mode = 1;
        else if (strcmp(flag, "-read") == 0) trace_mode = 2;
        else if (strcmp(flag, "-write") == 0) trace_mode = 3;
        else if (strcmp(flag, "-before") == 0) trace_mode = 4;
        else if (strcmp(flag, "-after") == 0) trace_mode = 5;
        else if (strcmp(flag, "-help") == 0) { display_help(); return 0; }
        else {
            printf("Unknown flag: %s\n", flag);
            display_help();
            return 1;
        }
    }

    init_cpu();

    int size = load_object(filename);
    if (size < 0) return 1;

    printf("Loaded %d instructions into memory\n", size);

    if (trace_mode == 4) { // -before
        dump_memory(size + 10);
    }

    run_cpu(0, trace_mode);

    if (trace_mode == 5) { // -after
        dump_memory(size + 10);
    }

    return 0;
}
