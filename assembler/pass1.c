/*
 * pass1.c - Pass 1: build symbol table, validate instructions and operands
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
#include "pass1.h"

/* The legendary PASS1 finally :< */
void pass1_build_symbols(ParsedLine program[], int count)
{
    int pc = 0;
    int i;
    Instruction *inst;
    char msg[256];
    int val;
    int idx;

    for(i = 0; i < count; i++)
    {
        program[i].address = pc;

        if(strlen(program[i].label) > 0)
        {
            add_symbol(program[i].label, pc, program[i].line_number);
        }

        if(strlen(program[i].mnemonic) > 0)
        {
            inst = find_instruction(program[i].mnemonic);

            if(inst == NULL)
            {
                sprintf(msg, "Error: invalid instruction '%s'", program[i].mnemonic);
                add_log_entry(msg, program[i].line_number, 1);
                printf("%s at line %d\n", msg, program[i].line_number);
            }
            else
            {
                if(inst->operand_type == 0 && strlen(program[i].operand) > 0)
                {
                    sprintf(msg, "Error: extra operand");
                    add_log_entry(msg, program[i].line_number, 1);
                    printf("%s at line %d\n", msg, program[i].line_number);
                }

                if(inst->operand_type != 0 && strlen(program[i].operand) == 0)
                {
                    sprintf(msg, "Error: missing operand");
                    add_log_entry(msg, program[i].line_number, 1);
                    printf("%s at line %d\n", msg, program[i].line_number);
                }

                if(strlen(program[i].operand) > 0 &&
                   !is_digit(program[i].operand[0]) &&
                   program[i].operand[0] != '-' &&
                   program[i].operand[0] != '+' &&
                   !is_hex(program[i].operand) &&
                   !is_octal(program[i].operand))
                {
                    record_reference(program[i].operand,
                                     program[i].line_number);
                }
            }

            if (strcmp(program[i].mnemonic, "SET") != 0) {
                pc++;
            }
            else if (strlen(program[i].label) > 0) {
                val = resolve_operand(program[i].operand, pc, 0);
                idx = find_symbol(program[i].label);
                if (idx != -1) {
                    symbols[idx].address = val;
                }
            }
        }
    }
}
