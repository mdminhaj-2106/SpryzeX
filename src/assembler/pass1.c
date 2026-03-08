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

    for(i = 0; i < count; i++)
    {
        program[i].address = pc;

        /* handle label */
        if(strlen(program[i].label) > 0)
        {
            add_symbol(program[i].label, pc, program[i].line_number);
        }

        /* validate instruction */
        if(strlen(program[i].mnemonic) > 0)
        {
            Instruction *inst = find_instruction(program[i].mnemonic);

            if(inst == NULL)
            {
                char msg[256];
                sprintf(msg, "Error: invalid instruction '%s'", program[i].mnemonic);
                add_log_entry(msg, program[i].line_number, 1);
                printf("%s at line %d\n", msg, program[i].line_number);
            }
            else
            {
                /* check operand rules */
                if(inst->operand_type == 0 && strlen(program[i].operand) > 0)
                {
                    char msg[256];
                    sprintf(msg, "Error: extra operand");
                    add_log_entry(msg, program[i].line_number, 1);
                    printf("%s at line %d\n", msg, program[i].line_number);
                }

                if(inst->operand_type != 0 && strlen(program[i].operand) == 0)
                {
                    char msg[256];
                    sprintf(msg, "Error: missing operand");
                    add_log_entry(msg, program[i].line_number, 1);
                    printf("%s at line %d\n", msg, program[i].line_number);
                }

                /* record label reference */
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

            /* SET is a directive that doesn't consume PC space in some designs, 
               but if it defines a label at a specific value, it's different.
               Actually, in this ISA, SET is like "label SET value".
               Let's assume it doesn't take space. */
            if (strcmp(program[i].mnemonic, "SET") != 0) {
                pc++;
            }
            else if (strlen(program[i].label) > 0) {
                /* Update symbol to the SET value instead of current PC */
                int val = resolve_operand(program[i].operand, pc, 0);
                /* Overwrite the symbol added earlier in the loop */
                int idx = find_symbol(program[i].label);
                if (idx != -1) {
                    symbols[idx].address = val;
                }
            }
        }
    }
}
