/*
 * parser.h - Parser interface
 * Author: Md Minhaj Uddin
 * Roll: 2401CS39
 * Declaration: I declare that this code is my own work.
 */
#ifndef PARSER_H
#define PARSER_H

#include "asm.h"

void parse_line(char *line, ParsedLine *out);
void print_symbol_table(void);

#endif
