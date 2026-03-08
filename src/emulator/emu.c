#include <stdio.h>
#include <stdlib.h>
#include "../assembler/state.h"
#include "cpu.h"

int memory[MEM_SIZE];
CPU cpu;

/* loads the .o file */
int load_object(const char *filename)
{
    FILE *file = fopen(filename, "rb");

    if(!file)
    {
        printf("Cannot open object file\n");
        return -1;
    }

    int count = 0;

    while(fread(&memory[count], sizeof(int), 1, file))
    {
        count++;
    }

    fclose(file);

    return count;
}


/* executes the instructions */
void execute(int opcode, int operand)
{
    switch(opcode)
    {
        case 0: /* ldc */
            cpu.B = cpu.A;
            cpu.A = operand;
            break;

        case 1: /* adc */
            cpu.A = cpu.A + operand;
            break;

        case 2: /* ldl */
            cpu.B = cpu.A;
            cpu.A = memory[cpu.SP + operand];
            break;

        case 3: /* stl */
            memory[cpu.SP + operand] = cpu.A;
            cpu.A = cpu.B;
            break;

        case 4: /* ldnl */
            cpu.A = memory[cpu.A + operand];
            break;

        case 5: /* stnl */
            memory[cpu.A + operand] = cpu.B;
            break;

        case 6: /* add */
            cpu.A = cpu.B + cpu.A;
            break;

        case 7: /* sub */
            cpu.A = cpu.B - cpu.A;
            break;

        case 8: /* shl */
            cpu.A = cpu.B << cpu.A;
            break;

        case 9: /* shr */
            cpu.A = cpu.B >> cpu.A;
            break;

        case 10: /* adj */
            cpu.SP = cpu.SP + operand;
            break;

        case 11: /* a2sp */
            cpu.SP = cpu.A;
            cpu.A = cpu.B;
            break;

        case 12: /* sp2a */
            cpu.B = cpu.A;
            cpu.A = cpu.SP;
            break;

        case 13: /* call */
            cpu.B = cpu.A;
            cpu.A = cpu.PC;
            cpu.PC = cpu.PC + operand;
            break;

        case 14: /* return */
            cpu.PC = cpu.A;
            cpu.A = cpu.B;
            break;

        case 15: /* brz */
            if(cpu.A == 0)
                cpu.PC = cpu.PC + operand;
            break;

        case 16: /* brlz */
            if(cpu.A < 0)
                cpu.PC = cpu.PC + operand;
            break;

        case 17: /* br */
            cpu.PC = cpu.PC + operand;
            break;

        case 18: /* HALT */
            printf("HALT reached\n");
            exit(0);

        default:
            printf("Unknown opcode %d\n", opcode);
            exit(1);
    }
}


/* execution loop */
void run_program()
{
    while(1)
    {
        int instruction = memory[cpu.PC++];

        int opcode = instruction & 0xFF;
        int operand = instruction >> 8;

        execute(opcode, operand);
    }
}

int main(int argc, char *argv[])
{
    if(argc != 2)
    {
        printf("Usage: %s program.o\n", argv[0]);
        return 1;
    }

    cpu.A = 0;
    cpu.B = 0;
    cpu.PC = 0;
    cpu.SP = 0;

    int size = load_object(argv[1]);

    printf("Loaded %d instructions\n", size);

    run_program();

    return 0;
}