/*
 * pass2.c - Pass 2: generate machine code, write .o, .lst, .log
 * Author: [YOUR FULL NAME]
 * User ID: [YOUR USER ID]
 * Declaration: I declare that this code is my own work.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "asm.h"
#include "state.h"
#include "utils.h"
#include "pass2.h"


/* Finally this is the Godly PASS2 */
void pass2_generate_code(ParsedLine program[], int count)
{
    int i;
    unsigned int code;
    machine_count = 0;

    for(i = 0; i < count; i++)
    {
        program[i].has_machine_code = 0;

        if(strlen(program[i].mnemonic) == 0)
            continue;

        code = encode_instruction(&program[i]);

        program[i].machine_code = code;
        program[i].has_machine_code = 1;

        if (strcmp(program[i].mnemonic, "SET") != 0) {
            machine_code[machine_count++] = code;
        } else {
            program[i].has_machine_code = 0; // Don't show in listing as machine code
        }
    }
}

/* Write onto an object/binary file */
void write_object_file(const char *filename)
{
    FILE *file = fopen(filename, "wb");

    if(file == NULL)
    {
        printf("Error creating object file: %s\n", filename);
        return;
    }

    int i;

    for(i = 0; i < machine_count; i++)
    {
        fwrite(&machine_code[i], sizeof(unsigned int), 1, file);
    }

    fclose(file);

    printf("Object file written successfully to %s\n", filename);
}

/* Write listing file (.lst) */
void write_listing_file(const char *filename, ParsedLine program[], int count)
{
    FILE *file = fopen(filename, "w");

    if(file == NULL)
    {
        printf("Error creating listing file: %s\n", filename);
        return;
    }

    int i;
    fprintf(file, "Line\tAddr\tMachine Code\tSource\n");
    fprintf(file, "----\t----\t------------\t------\n");

    for(i = 0; i < count; i++)
    {
        if(program[i].has_machine_code)
        {
            fprintf(file, "%d\t%04X\t%08X\t%s\n", 
                    program[i].line_number, 
                    program[i].address, 
                    program[i].machine_code, 
                    program[i].original_line);
        }
        else if(strlen(program[i].label) > 0)
        {
            fprintf(file, "%d\t%04X\t        \t%s\n", 
                    program[i].line_number, 
                    program[i].address, 
                    program[i].original_line);
        }
        else
        {
            fprintf(file, "%d\t    \t        \t%s\n", 
                    program[i].line_number, 
                    program[i].original_line);
        }
    }

    fclose(file);
    printf("Listing file written to %s\n", filename);
}

/* Write log file (.log) */
void write_log_file(const char *filename)
{
    FILE *file = fopen(filename, "w");

    if(file == NULL)
    {
        printf("Error creating log file: %s\n", filename);
        return;
    }

    int i;
    fprintf(file, "Assembler Log\n");
    fprintf(file, "=============\n\n");

    fprintf(file, "Symbol Table:\n");
    for(i = 0; i < symbol_count; i++)
    {
        fprintf(file, "%-20s %04X (line %d)\n", 
                symbols[i].label, 
                symbols[i].address, 
                symbols[i].line);
    }

    fprintf(file, "\nLabel References:\n");
    for(i = 0; i < reference_count; i++)
    {
        fprintf(file, "%-20s (line %d)\n",
                references[i].label,
                references[i].line);
    }

    fprintf(file, "\nErrors and Warnings:\n");
    if(log_count == 0)
    {
        fprintf(file, "No errors or warnings found.\n");
    }
    else
    {
        for(i = 0; i < log_count; i++)
        {
            fprintf(file, "[%s] Line %d: %s\n",
                    log_entries[i].is_error ? "ERROR" : "WARNING",
                    log_entries[i].line,
                    log_entries[i].message);
        }
    }

    fclose(file);
    printf("Log file written to %s\n", filename);
}
