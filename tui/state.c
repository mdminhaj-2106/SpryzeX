#include "state.h"

#include <stdarg.h>
#include <stdio.h>
#include <string.h>

static void copy_with_limit(char *dst, size_t dst_size, const char *src) {
    if (dst_size == 0) {
        return;
    }
    if (src == NULL) {
        dst[0] = '\0';
        return;
    }
    strncpy(dst, src, dst_size - 1);
    dst[dst_size - 1] = '\0';
}

void app_state_init(AppState *state) {
    memset(state, 0, sizeof(*state));

    state->editor.line_count = 1;
    state->editor.lines[0][0] = '\0';
    state->running = 1;
    state->output_mode = OUTPUT_MODE_LIVE;
    copy_with_limit(state->status, sizeof(state->status), "Ready");
}

void output_buffer_clear(OutputBuffer *buffer) {
    buffer->line_count = 0;
    buffer->scroll = 0;
}

void output_buffer_append_line(OutputBuffer *buffer, const char *line) {
    if (buffer->line_count >= SPRYZEX_MAX_OUTPUT_LINES) {
        int i;
        for (i = 1; i < SPRYZEX_MAX_OUTPUT_LINES; ++i) {
            memcpy(buffer->lines[i - 1], buffer->lines[i], SPRYZEX_MAX_OUTPUT_COLS);
        }
        buffer->line_count = SPRYZEX_MAX_OUTPUT_LINES - 1;
    }

    copy_with_limit(buffer->lines[buffer->line_count], SPRYZEX_MAX_OUTPUT_COLS, line);
    buffer->line_count++;
}

void output_buffer_append_text(OutputBuffer *buffer, const char *text) {
    const char *start;
    const char *end;
    char line[SPRYZEX_MAX_OUTPUT_COLS];
    size_t len;

    if (text == NULL || text[0] == '\0') {
        return;
    }

    start = text;
    while (*start != '\0') {
        end = strchr(start, '\n');
        if (end == NULL) {
            len = strlen(start);
        } else {
            len = (size_t)(end - start);
        }

        if (len >= sizeof(line)) {
            len = sizeof(line) - 1;
        }

        memcpy(line, start, len);
        line[len] = '\0';

        output_buffer_append_line(buffer, line);

        if (end == NULL) {
            break;
        }
        start = end + 1;
    }
}

void app_set_status(AppState *state, const char *fmt, ...) {
    va_list args;

    va_start(args, fmt);
    vsnprintf(state->status, sizeof(state->status), fmt, args);
    va_end(args);
}

const char *output_mode_name(OutputMode mode) {
    switch (mode) {
        case OUTPUT_MODE_LIVE:
            return "LIVE";
        case OUTPUT_MODE_LOG:
            return "LOG";
        case OUTPUT_MODE_LST:
            return "LST";
        case OUTPUT_MODE_OBJ:
            return "OBJ";
        default:
            return "?";
    }
}
