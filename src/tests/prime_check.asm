; Prime Check Test
; Check if 7 is prime (A = 1 if prime, 0 if not)

ldc 7
stl 0           ; number to check n = 7

ldc 2
stl 1           ; divisor i = 2

loop:
    ldl 0       ; n
    ldl 1       ; i
    sub         ; n - i
    brz is_prime ; if i == n, it is prime
    
    ; simplified modulo check: n - (n/i)*i
    ; but we don't have division.
    ; use subtraction loop for modulo.
    
    ldl 0       ; n
    stl 2       ; temp_n = n
    
mod_loop:
    ldl 2       ; temp_n
    ldl 1       ; i
    sub         ; A = temp_n - i
    brlz mod_done
    brz is_not_prime
    stl 2       ; temp_n = temp_n - i
    br mod_loop
    
mod_done:
    ldl 1       ; i
    adc 1       ; i++
    stl 1
    br loop

is_prime:
    ldc 1
    out
    HALT

is_not_prime:
    ldc 0
    out
    HALT
