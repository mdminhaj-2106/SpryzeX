; Loop Test: Calculates sum of numbers 1 to 10
; Sum should be 55 (0x37)

ldc 0           ; Initialize sum = 0
stl 0           ; sum at [SP+0]

ldc 10          ; Initialize counter = 10
stl 1
    
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
    out
    HALT
