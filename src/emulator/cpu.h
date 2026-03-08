#ifndef CPU_H
#define CPU_H

#define MEM_SIZE (1 << 24)

typedef struct {
    int A;
    int B;
    int PC;
    int SP;
} CPU;

extern int memory[MEM_SIZE];
extern CPU cpu;

/* cpu functions */
void init_cpu();
void execute_instruction(int instruction, int trace_mode);
void run_cpu(int limit, int trace_mode);

/* memory functions */
int load_object(const char *filename);
void dump_memory(int limit);

/* trace functions */
void trace_registers();
void trace_read(int address, int value);
void trace_write(int address, int old_val, int new_val);

#endif
