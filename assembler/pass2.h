/*
 * pass2.h - Pass 2 (code generation, listing, log) interface
 * Author: Md Minhaj Uddin
 * Roll: 2401CS39
 * Declaration: I declare that this code is my own work.
 */
#ifndef PASS2_H
#define PASS2_H

#include "asm.h"

void pass2_generate_code(ParsedLine program[], int count);
void write_object_file(const char *filename);
void write_listing_file(const char *filename, ParsedLine program[], int count);
void write_log_file(const char *filename);

#endif
