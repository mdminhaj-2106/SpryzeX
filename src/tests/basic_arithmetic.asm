; Basic Arithmetic Test
; Tests ldc, adc, add, sub, shl, shr
ldc 10
adc 5       ; A=15, B=10
ldc 20
add         ; A=35, B=20 (Wait, ldc 20: A=20, B=15. Then add: A=20+15=35)
ldc 10
sub         ; A=20-10=10 (ldc 10: A=10, B=35. Then sub: A=35-10=25)
ldc 2
shl         ; A=25 << 2 = 100
ldc 1
shr         ; A=100 >> 1 = 50
HALT
