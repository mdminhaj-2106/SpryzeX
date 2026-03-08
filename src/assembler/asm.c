#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include "asm.h"
#include "state.h"
#include "utils.h"
#include "parser.h"
#include "pass1.h"
#include "pass2.h"



/* Symbol Table */
Symbol symbols[MAX_LABELS];
int symbol_count = 0;

/* Label Reference */
LabelReference references[MAX_LINES];
int reference_count = 0;

/* Machine Code */
unsigned int machine_code[MAX_LINES];
int machine_count = 0;

/* Log Entries */
LogEntry log_entries[MAX_ERRORS];
int log_count = 0;
int error_occurred = 0;



/* instruction_table format */
Instruction instruction_table[] = {
    {"ldc", 0, 1},
    {"adc", 1, 1},
    {"ldl", 2, 1},
    {"stl", 3, 1},
    {"ldnl", 4, 1},
    {"stnl", 5, 1},
    {"add", 6, 0},
    {"sub", 7, 0},
    {"shl", 8, 0},
    {"shr", 9, 0},
    {"adj", 10, 1},
    {"a2sp", 11, 0},
    {"sp2a", 12, 0},
    {"call", 13, 2},
    {"return", 14, 0},
    {"brz", 15, 2},
    {"brlz", 16, 2},
    {"br", 17, 2},
    {"HALT", 18, 0},
    {"out", 21, 0},
    {"outc", 22, 0},
    {"data", 19, 1},
    {"SET", 20, 1}
};

int instruction_count = sizeof(instruction_table) / sizeof(Instruction);


int main(int argc, char *argv[])
{
    if(argc != 2)
    {
        printf("Usage: %s program.asm\n", argv[0]);
        return 1;
    }

    /* Ensure output directories exist */
    mkdir("outputs", 0777);
    mkdir("logs", 0777);
    mkdir("listings", 0777);

    char lines[MAX_LINES][MAX_LINE_LENGTH];
    ParsedLine program[MAX_LINES];

    int line_count = read_source(argv[1], lines);

    if(line_count < 0)
        return 1;

    printf("Read %d lines\n", line_count);

    int i;

    for(i = 0; i < line_count; i++)
    {
        program[i].line_number = i + 1;
        /* Copy original line before strtok modifies it */
        strcpy(program[i].original_line, lines[i]);
        /* remove newline if present */
        char *nl = strchr(program[i].original_line, '\n');
        if(nl) *nl = '\0';

        parse_line(lines[i], &program[i]);
    }
    
    printf("Parsing complete\n");
    
    pass1_build_symbols(program, line_count);

    check_undefined_labels();
    check_unused_labels();
    
    printf("Pass 1 complete\n");

    /* build output paths */
    char obj_path[256];
    char log_path[256];
    char lst_path[256];

    build_output_paths(argv[1], obj_path, log_path, lst_path);

    pass2_generate_code(program, line_count);
    
    if(!error_occurred)
    {
        write_object_file(obj_path);
    }
    else
    {
        printf("Assembly failed with errors. Object file NOT generated.\n");
    }

    write_listing_file(lst_path, program, line_count);
    write_log_file(log_path);

    printf("Pass 2 complete\n");

    return error_occurred;
}
