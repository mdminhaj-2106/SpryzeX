#ifndef UTILS_H
#define UTILS_H

#include "asm.h"

int read_source(const char *filename, char lines[][MAX_LINE_LENGTH]);
Instruction* find_instruction(char *name);
int find_symbol(char *label);
void record_reference(char *label, int line);
void add_symbol(char *label, int address, int line);
void check_undefined_labels();
void check_unused_labels();

#endif
