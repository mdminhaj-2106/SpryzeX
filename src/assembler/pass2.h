#ifndef PASS2_H
#define PASS2_H

#include "asm.h"

void pass2_generate_code(ParsedLine program[], int count);
void write_object_file(const char *filename);
void write_listing_file(const char *filename, ParsedLine program[], int count);
void write_log_file(const char *filename);

#endif
