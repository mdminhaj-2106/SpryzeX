; Loop Test: Calculates sum of numbers 1 to 10
; Sum should be 55 (0x37)

ldc 0           ; Initialize sum = 0
stl 0           ; sum at [SP+0]

ldc 10          ; Initialize counter = 10
loop:
    ldl 0       ; sum -> A, counter -> B
    add         ; sum = sum + counter
    stl 0       ; store new sum
    
    ldc -1
    adc 0       ; subtract 1 from counter? no, adc adds to A
    ; correct way to decrement B(counter):
    ; A is sum, B is counter
    ; load counter, sub 1, branch
    
    ldl 0       ; A = sum
    sp2a        ; A = SP, B = sum
    ldc 1       ; A = 1, B = SP
    add         ; A = SP+1
    ; wait, this ISA is tricky. 
    ; Let's re-do simple loop.
    
    ; SP+0: sum
    ; SP+1: counter
    
    ldc 0
    stl 0       ; sum=0
    ldc 10
    stl 1       ; counter=10
    
loop_start:
    ldl 1       ; A=counter
    brz done    ; if counter == 0, end
    
    ldl 0       ; A=sum, B=counter
    ldl 1       ; A=counter, B=sum
    add         ; A=sum+counter
    stl 0       ; sum = sum+counter, A=sum(old)
    
    ldl 1       ; A=counter
    adc -1      ; A=counter-1
    stl 1       ; counter = counter-1
    
    br loop_start
    
done:
    ldl 0       ; Final sum in A
    HALT
