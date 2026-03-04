#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "asm.h"



/* read_source reads .asm line-line and return stores it in lines[] */
int read_source(const char *filename, char lines[][MAX_LINE_LENGTH])
{
    FILE *file = fopen(filename, "r");

    if(file == NULL)
    {
        printf("Error opening file\n");
        return -1;
    }

    int count = 0;

    while(fgets(lines[count], MAX_LINE_LENGTH, file))
    {
        count++;
    }

    fclose(file);

    return count;
}


/* find_instruction for opcode lookup in instruction_table */
Instruction* find_instruction(char *name)
{
    int i;

    for(i = 0; i < instruction_count; i++)
    {
        if(strcmp(instruction_table[i].name, name) == 0)
        {
            return &instruction_table[i];
        }
    }

    return NULL;
}


/* find_symbol is basic SymbolTable lookup */
int find_symbol(char *label)
{
    int i;

    for(i = 0; i < symbol_count; i++)
    {
        if(strcmp(symbols[i].label, label) == 0)
        {
            return i;
        }
    }

    return -1;
}


/* record references to labels in pass1 */
void record_reference(char *label, int line)
{
    strcpy(references[reference_count].label, label);
    references[reference_count].line = line;
    reference_count++;
}


/* add_symbol inserts lables, address and line_number into SymbolTable */
void add_symbol(char *label, int address, int line)
{
    if(find_symbol(label) != -1)
    {
        printf("Error: duplicate label %s at line %d\n", label, line);
        return;
    }

    strcpy(symbols[symbol_count].label, label);
    symbols[symbol_count].address = address;
    symbols[symbol_count].line = line;

    symbol_count++;
}

/* Checking Undefined Labels */
void check_undefined_labels()
{
    int i;

    for(i = 0; i < reference_count; i++)
    {
        if(find_symbol(references[i].label) == -1)
        {
            printf("Error: label '%s' used at line %d but not defined\n",
                   references[i].label,
                   references[i].line);
        }
    }
}

/* Detect Unused Labels(Warning) */
void check_unused_labels()
{
    int i, j, used;

    for(i = 0; i < symbol_count; i++)
    {
        used = 0;

        for(j = 0; j < reference_count; j++)
        {
            if(strcmp(symbols[i].label, references[j].label) == 0)
            {
                used = 1;
                break;
            }
        }

        if(!used)
        {
            printf("Warning: label '%s' defined at line %d but never used\n",
                   symbols[i].label,
                   symbols[i].line);
        }
    }
}