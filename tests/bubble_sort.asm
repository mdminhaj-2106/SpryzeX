; Bubble Sort Test (Natural Human Essence style)
; Sort an array of 5 elements: 5, 4, 3, 2, 1

ldc 0
stl 0           ; outer loop counter i = 0

outer_loop:
    ldc 0
    stl 1       ; inner loop counter j = 0

inner_loop:
    ; compare array[j] and array[j+1]
    ldc arr
    ldl 1       ; j
    add         ; A = arr+j
    ldnl 0      ; A = arr[j]
    stl 2       ; temp1 = arr[j]
    
    ldc arr
    ldl 1
    add         ; A = arr + j
    adc 1       ; A = arr+j+1
    ldnl 0      ; A = arr[j+1]
    
    ldl 2       ; A = arr[j], B = arr[j+1]
    sub         ; A = B - A = arr[j+1] - arr[j]
    brlz swap   ; if arr[j+1] < arr[j], swap
    br next_j

swap:
    ; arr[j] is in memory[addr_j], arr[j+1] is in memory[addr_j+1]
    ldc arr
    ldl 1       ; j
    add         ; A = addr_j
    stl 3       ; SP+3 = addr_j
    
    ldc arr
    ldl 1
    add         ; A = arr + j
    adc 1       ; A = addr_j+1
    stl 4       ; SP+4 = addr_j+1
    
    ldl 3
    ldnl 0      ; A = arr[j]
    stl 5       ; SP+5 = temp = arr[j]
    
    ldl 4
    ldnl 0      ; A = arr[j+1]
    ldl 3       ; A = addr_j, B = arr[j+1]
    stnl 0      ; memory[addr_j] = arr[j+1]
    
    ldl 5       ; A = temp
    ldl 4       ; A = addr_j+1, B = temp
    stnl 0      ; memory[addr_j+1] = temp

next_j:
    ldl 1       ; j
    adc 1
    stl 1
    ldc 4       ; j < 4? (array size - 1, avoids arr[j+1] out-of-bounds)
    ldl 1
    sub         ; A = 4 - j
    brlz inner_done
    brz inner_done
    br inner_loop

inner_done:
    ldl 0       ; i
    adc 1
    stl 0
    ldc 4       ; i < 4? (array size - 1 passes)
    ldl 0
    sub
    brlz outer_done
    brz outer_done
    br outer_loop

outer_done:
    ; print sorted array (one per line)
    ldc arr
    ldnl 0
    out
    ldc arr
    ldnl 1
    out
    ldc arr
    ldnl 2
    out
    ldc arr
    ldnl 3
    out
    ldc arr
    ldnl 4
    out
    HALT

arr:
    data 5
    data 4
    data 3
    data 2
    data 1