; ─────────────────────────────────────────────────────
;  SPRYZEX IDE  —  Hello World Sample
;  Press Ctrl+B to assemble, Ctrl+R to run
; ─────────────────────────────────────────────────────

start:
        ldc  72          ; H
        outc
        ldc  101         ; e
        outc
        ldc  108         ; l
        outc
        ldc  108         ; l
        outc
        ldc  111         ; o
        outc
        ldc  44          ; ,
        outc
        ldc  32          ; space
        outc
        ldc  87          ; W
        outc
        ldc  111         ; o
        outc
        ldc  114         ; r
        outc
        ldc  108         ; l
        outc
        ldc  100         ; d
        outc
        ldc  33          ; !
        outc
        ldc  10          ; newline
        outc
        HALT
