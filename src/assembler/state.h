#ifndef ASSEMBLER_STATE_H
#define ASSEMBLER_STATE_H

#include "asm.h"

/* Shared assembler state (defined in asm.c). */
extern Symbol symbols[MAX_LABELS];
extern int symbol_count;

extern LabelReference references[1000];
extern int reference_count;

extern Instruction instruction_table[];
extern int instruction_count;

#endif

