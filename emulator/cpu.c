/*
 * cpu.c - CPU init, instruction execution, run loop
 * Author: [YOUR FULL NAME]
 * User ID: [YOUR USER ID]
 * Declaration: I declare that this code is my own work.
 */
#include <stdio.h>
#include <stdlib.h>
#include "cpu.h"

CPU cpu;

/* humble cpu initialization */
void init_cpu() {
    cpu.A = 0;
    cpu.B = 0;
    cpu.PC = 0;
    cpu.SP = 10000; /* stack far away from code */
}

/* execution logic for each opcode */
void execute_instruction(int instruction, int trace_mode) {
    int opcode = instruction & 0xFF;
    int operand = instruction >> 8;
    
    /* sign extend 24-bit operand to 32-bit int */
    if (operand & 0x800000) {
        operand |= 0xFF000000;
    }

    if (trace_mode == 1) { // -trace flag
        trace_registers();
    }
    
    switch(opcode)
    {
        case 0: /* ldc */
            cpu.B = cpu.A;
            cpu.A = operand;
            if (trace_mode == 1) printf("ldc %d\n", operand);
            break;

        case 1: /* adc */
            cpu.A = cpu.A + operand;
            if (trace_mode == 1) printf("adc %d\n", operand);
            break;

        case 2: /* ldl */
            cpu.B = cpu.A;
            cpu.A = memory[cpu.SP + operand];
            if (trace_mode == 2) trace_read(cpu.SP + operand, cpu.A);
            if (trace_mode == 1) printf("ldl %d\n", operand);
            break;

        case 3: /* stl */
            {
                int old = memory[cpu.SP + operand];
                memory[cpu.SP + operand] = cpu.A;
                if (trace_mode == 3) trace_write(cpu.SP + operand, old, cpu.A);
                cpu.A = cpu.B;
                if (trace_mode == 1) printf("stl %d\n", operand);
            }
            break;

        case 4: /* ldnl */
            {
                int address = cpu.A + operand;
                cpu.A = memory[address];
                if (trace_mode == 2) trace_read(address, cpu.A);
            }
            if (trace_mode == 1) printf("ldnl %d\n", operand);
            break;

        case 5: /* stnl */
            {
                int old = memory[cpu.A + operand];
                memory[cpu.A + operand] = cpu.B;
                if (trace_mode == 3) trace_write(cpu.A + operand, old, cpu.B);
                if (trace_mode == 1) printf("stnl %d\n", operand);
            }
            break;

        case 6: /* add */
            cpu.A = cpu.B + cpu.A;
            if (trace_mode == 1) printf("add\n");
            break;

        case 7: /* sub */
            cpu.A = cpu.B - cpu.A;
            if (trace_mode == 1) printf("sub\n");
            break;

        case 8: /* shl */
            cpu.A = cpu.B << cpu.A;
            if (trace_mode == 1) printf("shl\n");
            break;

        case 9: /* shr */
            cpu.A = cpu.B >> cpu.A;
            if (trace_mode == 1) printf("shr\n");
            break;

        case 10: /* adj */
            cpu.SP = cpu.SP + operand;
            if (trace_mode == 1) printf("adj %d\n", operand);
            break;

        case 11: /* a2sp */
            cpu.SP = cpu.A;
            cpu.A = cpu.B;
            if (trace_mode == 1) printf("a2sp\n");
            break;

        case 12: /* sp2a */
            cpu.B = cpu.A;
            cpu.A = cpu.SP;
            if (trace_mode == 1) printf("sp2a\n");
            break;

        case 13: /* call */
            cpu.B = cpu.A;
            cpu.A = cpu.PC;
            cpu.PC = cpu.PC + operand;
            if (trace_mode == 1) printf("call %d\n", operand);
            break;

        case 14: /* return */
            cpu.PC = cpu.A;
            cpu.A = cpu.B;
            if (trace_mode == 1) printf("return\n");
            break;

        case 15: /* brz */
            if(cpu.A == 0)
                cpu.PC = cpu.PC + operand;
            if (trace_mode == 1) printf("brz %d\n", operand);
            break;

        case 16: /* brlz */
            if(cpu.A < 0)
                cpu.PC = cpu.PC + operand;
            if (trace_mode == 1) printf("brlz %d\n", operand);
            break;

        case 17: /* br */
            cpu.PC = cpu.PC + operand;
            if (trace_mode == 1) printf("br %d\n", operand);
            break;

        case 18: /* HALT */
            if (trace_mode == 1) printf("HALT\n");
            printf("HALT reached at PC=%08X\n", cpu.PC - 1);
            return;

        case 21: /* out */
            if (trace_mode == 1) printf("out\n");
            printf("%d\n", cpu.A);
            fflush(stdout);
            break;

        case 22: /* outc */
            if (trace_mode == 1) printf("outc\n");
            putchar(cpu.A & 0xFF);
            fflush(stdout);
            break;

        default:
            printf("Error: Unknown opcode %d at PC=%08X\n", opcode, cpu.PC - 1);
            exit(1);
    }
}

/* execution loop */
void run_cpu(int limit, int trace_mode) {
    int instructions_executed = 0;
    while(1) {
        int instruction = memory[cpu.PC++];
        int opcode = instruction & 0xFF;
        execute_instruction(instruction, trace_mode);
        instructions_executed++;

        if (opcode == 18) break; /* HALT */

        if (limit > 0 && instructions_executed >= limit) {
            printf("Execution limit reached\n");
            break;
        }

        if (instructions_executed > (1 << 24)) {
            printf("Error: Infinite loop detected or memory exceeded\n");
            exit(1);
        }
    }
}
