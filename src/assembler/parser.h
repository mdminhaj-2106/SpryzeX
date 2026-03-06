#ifndef PARSER_H
#define PARSER_H

#include "asm.h"

void parse_line(char *line, ParsedLine *out);
void print_symbol_table();

#endif
