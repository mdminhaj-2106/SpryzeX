#include <stdio.h>
#include <stdlib.h>
#include "cpu.h"

int memory[MEM_SIZE];

/* loading object file into memory */
int load_object(const char *filename) {
    FILE *file = fopen(filename, "rb");
    if (!file) {
        printf("Error: Could not open file %s\n", filename);
        return -1;
    }

    int count = 0;
    while (fread(&memory[count], sizeof(int), 1, file)) {
        count++;
        if (count >= MEM_SIZE) {
            printf("Warning: Object file too large for memory\n");
            break;
        }
    }

    fclose(file);
    return count;
}

/* dumping memory contents */
void dump_memory(int limit) {
    printf("\nMemory Dump (up to %d):\n", limit);
    for (int i = 0; i < limit; i += 4) {
        printf("%08X: %08X %08X %08X %08X\n", i, memory[i], memory[i+1], memory[i+2], memory[i+3]);
    }
}
