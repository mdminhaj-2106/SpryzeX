/*
 * trace.c - Register and memory trace output for debug
 * Author: [YOUR FULL NAME]
 * User ID: [YOUR USER ID]
 * Declaration: I declare that this code is my own work.
 */
#include <stdio.h>
#include "cpu.h"

/* Registers trace */
void trace_registers() {
    printf("PC: %08X SP: %08X A: %08X B: %08X ", cpu.PC, cpu.SP, cpu.A, cpu.B);
}

/* Reading memory trace */
void trace_read(int address, int value) {
    printf("Read memory[%08X] -> %08X\n", address, value);
}

/* Writing memory trace */
void trace_write(int address, int old_val, int new_val) {
    printf("Write memory[%08X]: was %08X now %08X\n", address, old_val, new_val);
}
