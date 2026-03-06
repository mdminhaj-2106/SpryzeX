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
                printf("Error: invalid instruction '%s' at line %d\n",
                       program[i].mnemonic,
                       program[i].line_number);
            }
            else
            {
                /* check operand rules */
                if(inst->operand_type == 0 && strlen(program[i].operand) > 0)
                {
                    printf("Error: extra operand at line %d\n",
                           program[i].line_number);
                }

                if(inst->operand_type != 0 && strlen(program[i].operand) == 0)
                {
                    printf("Error: missing operand at line %d\n",
                           program[i].line_number);
                }

                /* record label reference */
                if(strlen(program[i].operand) > 0 &&
                   !isdigit(program[i].operand[0]))
                {
                    record_reference(program[i].operand,
                                     program[i].line_number);
                }
            }

            pc++;
        }
    }
}
