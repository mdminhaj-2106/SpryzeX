#ifndef ASM_H
#define ASM_H

#define MAX_LINES 1000
#define MAX_LABELS 500
#define MAX_LINE_LENGTH 256


/* instruction information */
typedef struct {
    char name[10];
    int opcode;
    int operand_type;
} Instruction;

/* symbol table entry */
typedef struct {
    char label[50];
    int address;
    int line;
} Symbol;

/* for later filling of symbol table */
typedef struct {
    char label[50];
    int line;
} LabelReference;

/* parsed line structure */
typedef struct {
    int address;
    char label[50];
    char mnemonic[20];
    char operand[50];
    int line_number;
} ParsedLine;

#endif