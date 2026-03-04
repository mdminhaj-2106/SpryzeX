#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "asm.h"
#include "utils.h"



/* Symbol Table */
Symbol symbols[MAX_LABELS];
int symbol_count = 0;

/* Label Reference */
LabelReference references[1000];
int reference_count = 0;



/* instruction_table format */
Instruction instruction_table[] = {
    {"ldc", 0, 1},
    {"adc", 1, 1},
    {"ldl", 2, 2},
    {"stl", 3, 2},
    {"ldnl", 4, 2},
    {"stnl", 5, 2},
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
    {"HALT", 18, 0}
};

int instruction_count = sizeof(instruction_table) / sizeof(Instruction);


int main(int argc, char *argv[])
{
    if(argc != 2)
    {
        printf("Usage: %s program.asm\n", argv[0]);
        return 1;
    }

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
        parse_line(lines[i], &program[i]);
    }
    
    printf("Parsing complete\n");
    
    pass1_build_symbols(program, line_count);
    print_symbol_table();

    check_undefined_labels();
    check_unused_labels();
    
    printf("Pass1 Build complete\n");

    return 0;
}