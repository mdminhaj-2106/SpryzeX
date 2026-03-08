#ifndef UTILS_H
#define UTILS_H

#include "asm.h"

int read_source(const char *filename, char lines[][MAX_LINE_LENGTH]);
Instruction* find_instruction(char *name);
int find_symbol(char *label);
void add_symbol(char *label, int address, int line);
void record_reference(char *label, int line);
void check_undefined_labels();
void check_unused_labels();
void add_log_entry(const char *message, int line, int is_error);
int resolve_operand(char *operand, int current_address, int is_branch);
unsigned int encode_instruction(ParsedLine *line);
void build_output_paths(char *input, char *obj, char *log, char *lst);

/* Base conversion/validation helpers */
int is_digit(char c);
int is_valid_label(const char *label);
int is_hex(const char *s);
int is_octal(const char *s);
int is_decimal(const char *s);

#endif
