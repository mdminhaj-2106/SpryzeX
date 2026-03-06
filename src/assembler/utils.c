#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "asm.h"
#include "state.h"
#include "utils.h"

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
        char msg[256];
        sprintf(msg, "Error: duplicate label %s", label);
        add_log_entry(msg, line, 1);
        printf("%s at line %d\n", msg, line);
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
            char msg[256];
            sprintf(msg, "Error: label '%s' used but not defined", references[i].label);
            add_log_entry(msg, references[i].line, 1);
            printf("%s at line %d\n", msg, references[i].line);
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
            char msg[256];
            sprintf(msg, "Warning: label '%s' defined but never used", symbols[i].label);
            add_log_entry(msg, symbols[i].line, 0);
            printf("%s at line %d\n", msg, symbols[i].line);
        }
    }
}

void add_log_entry(const char *message, int line, int is_error)
{
    if(log_count < MAX_ERRORS)
    {
        strncpy(log_entries[log_count].message, message, 255);
        log_entries[log_count].line = line;
        log_entries[log_count].is_error = is_error;
        log_count++;
        
        if(is_error)
            error_occurred = 1;
    }
}

/* Converts the operand text -> Integer */
int resolve_operand(char *operand, int current_address, int is_branch)
{
    int value;

    /* if operand is numeric */
    if(isdigit(operand[0]) || operand[0] == '-' || operand[0] == '+')
    {
        value = atoi(operand);
    }
    else
    {
        int idx = find_symbol(operand);

        if(idx == -1)
        {
            char msg[256];
            sprintf(msg, "Error: undefined label %s", operand);
            add_log_entry(msg, -1, 1); // Line unknown here, should be passed or handled
            printf("%s\n", msg);
            return 0;
        }

        value = symbols[idx].address;

        if(is_branch)
        {
            value = value - (current_address + 1);
        }
    }

    return value;
}


/* Encode the instruction into machine_code */
unsigned int encode_instruction(ParsedLine *line)
{
    Instruction *inst = find_instruction(line->mnemonic);

    if(inst == NULL)
        return 0;

    int operand = 0;

    if(inst->operand_type != 0)
    {
        int is_branch = (inst->operand_type == 2);
        operand = resolve_operand(line->operand, line->address, is_branch);
    }

    unsigned int code = (operand << 8) | inst->opcode;

    return code;
}

/* Helper for writing .o, .log and .lst files */
void build_output_paths(char *input,
                        char *obj,
                        char *log,
                        char *lst)
{
    char base[256];
    char *dot;

    /* copy input path */
    strcpy(base, input);

    /* remove directory path if present */
    char *slash = strrchr(base, '/');
    if(slash != NULL)
        memmove(base, slash + 1, strlen(slash));

    /* remove extension */
    dot = strrchr(base, '.');
    if(dot)
        *dot = '\0';

    /* build output paths */
    sprintf(obj, "outputs/%s.o", base);
    sprintf(log, "logs/%s.log", base);
    sprintf(lst, "listings/%s.lst", base);
}
