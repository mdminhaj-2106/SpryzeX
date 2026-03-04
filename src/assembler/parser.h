#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "asm.h"
#include "utils.h"

/* the heavenly parser who parses a line and stores it in ParsedLine */
void parse_line(char *line, ParsedLine *out)
{
    char *token;

    out->label[0] = '\0';
    out->mnemonic[0] = '\0';
    out->operand[0] = '\0';

    token = strtok(line, " \t\n");

    if(token == NULL)
        return;

    /* check if token is label */
    if(token[strlen(token)-1] == ':')
    {
        token[strlen(token)-1] = '\0';
        strcpy(out->label, token);

        token = strtok(NULL, " \t\n");
    }

    if(token != NULL)
    {
        strcpy(out->mnemonic, token);
        token = strtok(NULL, " \t\n");
    }

    if(token != NULL)
    {
        strcpy(out->operand, token);
    }
}


/* logging SymbolTable onto terminal for debugging */
void print_symbol_table()
{
    int i;

    printf("\nSymbol Table:\n");

    for(i = 0; i < symbol_count; i++)
    {
        printf("%s -> %d\n", symbols[i].label, symbols[i].address);
    }
}
