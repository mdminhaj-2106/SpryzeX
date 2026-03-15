; Success: Demonstrates SET directive. "label: SET value" defines a symbol without
; consuming a word of code. We use max: SET 100 then output 100.
max: SET 100
    ldc max
    out
    HALT
