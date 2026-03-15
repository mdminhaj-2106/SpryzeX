/*
 * state.h - Shared assembler state (extern declarations)
 * Author: [YOUR FULL NAME]
 * User ID: [YOUR USER ID]
 * Declaration: I declare that this code is my own work.
 */
#ifndef ASSEMBLER_STATE_H
#define ASSEMBLER_STATE_H

#include "asm.h"

/* Shared assembler state (defined in asm.c). */
extern Symbol symbols[MAX_LABELS];
extern int symbol_count;

extern LabelReference references[MAX_LINES];
extern int reference_count;

extern Instruction instruction_table[];
extern int instruction_count;

extern unsigned int machine_code[MAX_LINES];
extern int machine_count;

extern LogEntry log_entries[MAX_ERRORS];
extern int log_count;
extern int error_occurred;

#endif
