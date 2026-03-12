#include <stdio.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/time.h>

#include <ncurses.h>

#include "editor.h"
#include "mascot.h"
#include "runner.h"
#include "state.h"
#include "ui.h"

static long long now_ms(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (long long)tv.tv_sec * 1000LL + (long long)(tv.tv_usec / 1000);
}

static void ensure_samples_dir(void) {
    mkdir("samples", 0777);
}

static void set_default_file(AppState *state, int argc, char **argv) {
    if (argc > 1 && argv[1] != NULL && argv[1][0] != '\0') {
        strncpy(state->current_file, argv[1], sizeof(state->current_file) - 1);
        state->current_file[sizeof(state->current_file) - 1] = '\0';
        return;
    }

    ensure_samples_dir();
    strncpy(state->current_file, "samples/code.asm", sizeof(state->current_file) - 1);
    state->current_file[sizeof(state->current_file) - 1] = '\0';
}

static void seed_default_content_if_empty(AppState *state) {
    if (state->editor.line_count == 1 && state->editor.lines[0][0] == '\0') {
        strncpy(state->editor.lines[0], "        ldc 72", SPRYZEX_MAX_COLS - 1);
        strncpy(state->editor.lines[1], "        outc", SPRYZEX_MAX_COLS - 1);
        strncpy(state->editor.lines[2], "        ldc 105", SPRYZEX_MAX_COLS - 1);
        strncpy(state->editor.lines[3], "        outc", SPRYZEX_MAX_COLS - 1);
        strncpy(state->editor.lines[4], "        HALT", SPRYZEX_MAX_COLS - 1);

        state->editor.lines[0][SPRYZEX_MAX_COLS - 1] = '\0';
        state->editor.lines[1][SPRYZEX_MAX_COLS - 1] = '\0';
        state->editor.lines[2][SPRYZEX_MAX_COLS - 1] = '\0';
        state->editor.lines[3][SPRYZEX_MAX_COLS - 1] = '\0';
        state->editor.lines[4][SPRYZEX_MAX_COLS - 1] = '\0';
        state->editor.line_count = 5;
        state->editor.cursor_row = 0;
        state->editor.cursor_col = 0;
        state->editor.dirty = 1;
    }
}

static void load_initial_file(AppState *state) {
    char err[256];

    if (editor_load_file(&state->editor, state->current_file, err, sizeof(err)) != 0) {
        app_set_status(state, "New file: %s", state->current_file);
        seed_default_content_if_empty(state);
        return;
    }

    app_set_status(state, "Loaded %s", state->current_file);
}

static void save_current(AppState *state) {
    char err[256];

    if (editor_save_file(&state->editor, state->current_file, err, sizeof(err)) != 0) {
        app_set_status(state, "%s", err);
        return;
    }

    runner_update_artifact_paths(state);
    app_set_status(state, "Saved %s", state->current_file);
}

static OutputBuffer *active_output(AppState *state) {
    if (state->output_mode == OUTPUT_MODE_LIVE) {
        return &state->live_output;
    }
    return &state->artifact_output;
}

static void scroll_output(AppState *state, int delta, int view_rows) {
    OutputBuffer *buffer = active_output(state);
    int max_scroll = buffer->line_count - view_rows;

    if (max_scroll < 0) {
        max_scroll = 0;
    }

    buffer->scroll += delta;
    if (buffer->scroll < 0) {
        buffer->scroll = 0;
    }
    if (buffer->scroll > max_scroll) {
        buffer->scroll = max_scroll;
    }
}

static void open_file_prompt(UIContext *ui, AppState *state) {
    char path[PATH_MAX];
    char err[256];

    if (state->editor.dirty) {
        app_set_status(state, "Unsaved changes. Save with S before opening another file.");
        return;
    }

    if (ui_prompt_input(ui, "Open file:", path, sizeof(path)) != 0) {
        app_set_status(state, "Open canceled");
        return;
    }

    strncpy(state->current_file, path, sizeof(state->current_file) - 1);
    state->current_file[sizeof(state->current_file) - 1] = '\0';
    runner_update_artifact_paths(state);

    if (editor_load_file(&state->editor, state->current_file, err, sizeof(err)) != 0) {
        app_set_status(state, "New file: %s", state->current_file);
        return;
    }

    app_set_status(state, "Loaded %s", state->current_file);
}

static void handle_key(UIContext *ui, AppState *state, int ch) {
    if (ch == ERR) {
        return;
    }

    if (ch == KEY_RESIZE) {
        ui_handle_resize(ui);
        editor_ensure_cursor_visible(&state->editor, ui->editor_view_rows, ui->editor_view_cols);
        return;
    }

    switch (ch) {
        case 'Q':
            state->running = 0;
            return;

        case 'S':
            save_current(state);
            return;

        case 'B':
            save_current(state);
            runner_build(state);
            return;

        case 'R':
            runner_run_emulator(state);
            return;

        case 'O':
            open_file_prompt(ui, state);
            return;

        case '[':
            scroll_output(state, 1, ui->output_view_rows);
            return;

        case ']':
            scroll_output(state, -1, ui->output_view_rows);
            return;

        case '\t': {
            OutputMode next = (OutputMode)(((int)state->output_mode + 1) % 4);
            runner_set_output_mode(state, next);
            app_set_status(state, "Output view: %s", output_mode_name(next));
            return;
        }

        case KEY_F(2):
            runner_set_output_mode(state, OUTPUT_MODE_LIVE);
            app_set_status(state, "Output view: LIVE");
            return;

        case KEY_F(3):
            runner_set_output_mode(state, OUTPUT_MODE_LOG);
            app_set_status(state, "Output view: LOG");
            return;

        case KEY_F(4):
            runner_set_output_mode(state, OUTPUT_MODE_LST);
            app_set_status(state, "Output view: LST");
            return;

        case KEY_F(5):
            runner_set_output_mode(state, OUTPUT_MODE_OBJ);
            app_set_status(state, "Output view: OBJ");
            return;

        default:
            break;
    }

    if (editor_handle_key(&state->editor, ch)) {
        editor_ensure_cursor_visible(&state->editor, ui->editor_view_rows, ui->editor_view_cols);
    }
}

int main(int argc, char **argv) {
    UIContext ui;
    AppState state;

    app_state_init(&state);
    set_default_file(&state, argc, argv);
    runner_update_artifact_paths(&state);
    load_initial_file(&state);

    if (ui_init(&ui) != 0) {
        endwin();
        fprintf(stderr, "Failed to initialize ncurses UI. Make sure terminal size is sufficient.\n");
        return 1;
    }

    while (state.running) {
        int ch;

        mascot_tick(&state, now_ms());
        ui_render(&ui, &state);

        ch = getch();
        handle_key(&ui, &state, ch);
    }

    ui_destroy(&ui);
    return 0;
}
