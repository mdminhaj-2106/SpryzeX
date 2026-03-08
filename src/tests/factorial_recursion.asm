; Factorial Loop Test: factorial(5) = 120 (0x78)

ldc 5
stl 0           ; n = 5
ldc 1
stl 1           ; result = 1

fact_loop:
    ldl 0       ; A = n
    brz done    ; if n == 0, done
    
    ; multiplication result = result * n
    ldl 1       ; A = result
    stl 2       ; temp_res = result
    ldc 0
    stl 3       ; new_res = 0
    ldl 0       ; A = n
    stl 4       ; count = n
    
mult_loop:
    ldl 4       ; count
    brz mult_done
    
    ldl 3       ; new_res
    ldl 2       ; temp_res
    add         ; new_res + temp_res
    stl 3
    
    ldl 4       ; count
    adc -1      ; count--
    stl 4
    br mult_loop
    
mult_done:
    ldl 3       ; A = new_res
    stl 1       ; result = new_res
    
    ldl 0       ; A = n
    adc -1      ; n--
    stl 0
    br fact_loop

done:
    ldl 1       ; A = Final result
    HALT
