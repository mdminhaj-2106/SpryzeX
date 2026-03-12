#ifndef SPRYZEX_STATE_H
#define SPRYZEX_STATE_H

#include <limits.h>
#include <stddef.h>

#define SPRYZEX_MAX_LINES 2000
#define SPRYZEX_MAX_COLS 256
#define SPRYZEX_MAX_OUTPUT_LINES 4000
#define SPRYZEX_MAX_OUTPUT_COLS 512
#define SPRYZEX_STATUS_LEN 256

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif

typedef enum OutputMode {
    OUTPUT_MODE_LIVE = 0,
    OUTPUT_MODE_LOG = 1,
    OUTPUT_MODE_LST = 2,
    OUTPUT_MODE_OBJ = 3
} OutputMode;

typedef struct EditorState {
    char lines[SPRYZEX_MAX_LINES][SPRYZEX_MAX_COLS];
    int line_count;
    int cursor_row;
    int cursor_col;
    int row_offset;
    int col_offset;
    int dirty;
} EditorState;

typedef struct OutputBuffer {
    char lines[SPRYZEX_MAX_OUTPUT_LINES][SPRYZEX_MAX_OUTPUT_COLS];
    int line_count;
    int scroll;
} OutputBuffer;

typedef struct AppState {
    EditorState editor;
    OutputBuffer live_output;
    OutputBuffer artifact_output;

    char current_file[PATH_MAX];
    char obj_path[PATH_MAX];
    char log_path[PATH_MAX];
    char lst_path[PATH_MAX];

    OutputMode output_mode;
    int running;

    int mascot_frame;
    long long last_mascot_tick_ms;

    char status[SPRYZEX_STATUS_LEN];
} AppState;

void app_state_init(AppState *state);

void output_buffer_clear(OutputBuffer *buffer);
void output_buffer_append_line(OutputBuffer *buffer, const char *line);
void output_buffer_append_text(OutputBuffer *buffer, const char *text);

void app_set_status(AppState *state, const char *fmt, ...);

const char *output_mode_name(OutputMode mode);

#endif
