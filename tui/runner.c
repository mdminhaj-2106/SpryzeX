#include "runner.h"

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>

static void shell_quote(const char *src, char *dst, size_t dst_size) {
    size_t i;
    size_t j;

    if (dst_size == 0) {
        return;
    }

    j = 0;
    dst[j++] = '\'';

    for (i = 0; src[i] != '\0' && j + 5 < dst_size; ++i) {
        if (src[i] == '\'') {
            dst[j++] = '\'';
            dst[j++] = '\\';
            dst[j++] = '\'';
            dst[j++] = '\'';
        } else {
            dst[j++] = src[i];
        }
    }

    if (j < dst_size - 1) {
        dst[j++] = '\'';
    }

    dst[j] = '\0';
}

static void ensure_output_dirs(void) {
    mkdir("outputs", 0777);
    mkdir("logs", 0777);
    mkdir("listings", 0777);
}

static int run_command_capture(AppState *state, const char *cmd, const char *label) {
    FILE *pipe;
    char line[SPRYZEX_MAX_OUTPUT_COLS];
    int status;

    output_buffer_clear(&state->live_output);
    output_buffer_append_line(&state->live_output, label);
    output_buffer_append_line(&state->live_output, cmd);
    output_buffer_append_line(&state->live_output, "");

    pipe = popen(cmd, "r");
    if (pipe == NULL) {
        app_set_status(state, "Failed to run command: %s", strerror(errno));
        output_buffer_append_line(&state->live_output, "popen failed");
        state->output_mode = OUTPUT_MODE_LIVE;
        return -1;
    }

    while (fgets(line, sizeof(line), pipe) != NULL) {
        output_buffer_append_text(&state->live_output, line);
    }

    status = pclose(pipe);
    state->output_mode = OUTPUT_MODE_LIVE;

    if (status == -1) {
        app_set_status(state, "%s finished with pclose error", label);
        return -1;
    }

    if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
        app_set_status(state, "%s: success", label);
        return 0;
    }

    if (WIFEXITED(status)) {
        app_set_status(state, "%s: failed (exit %d)", label, WEXITSTATUS(status));
    } else {
        app_set_status(state, "%s: failed", label);
    }

    return -1;
}

void runner_update_artifact_paths(AppState *state) {
    char base[PATH_MAX];
    const char *filename;
    char *dot;

    filename = strrchr(state->current_file, '/');
    if (filename == NULL) {
        filename = state->current_file;
    } else {
        filename++;
    }

    strncpy(base, filename, sizeof(base) - 1);
    base[sizeof(base) - 1] = '\0';

    dot = strrchr(base, '.');
    if (dot != NULL) {
        *dot = '\0';
    }

    snprintf(state->obj_path, sizeof(state->obj_path), "outputs/%s.o", base);
    snprintf(state->log_path, sizeof(state->log_path), "logs/%s.log", base);
    snprintf(state->lst_path, sizeof(state->lst_path), "listings/%s.lst", base);
}

int runner_build(AppState *state) {
    char quoted_file[PATH_MAX * 2];
    char cmd[PATH_MAX * 3];
    int rc;

    ensure_output_dirs();
    runner_update_artifact_paths(state);

    shell_quote(state->current_file, quoted_file, sizeof(quoted_file));
    snprintf(cmd, sizeof(cmd), "./asm %s 2>&1", quoted_file);

    rc = run_command_capture(state, cmd, "Build");

    if (state->output_mode != OUTPUT_MODE_LIVE) {
        runner_load_selected_artifact(state);
    }

    return rc;
}

int runner_run_emulator(AppState *state) {
    char quoted_obj[PATH_MAX * 2];
    char cmd[PATH_MAX * 3];

    runner_update_artifact_paths(state);

    shell_quote(state->obj_path, quoted_obj, sizeof(quoted_obj));
    snprintf(cmd, sizeof(cmd), "./emu %s 2>&1", quoted_obj);

    return run_command_capture(state, cmd, "Run");
}

static int load_text_file(OutputBuffer *buffer, const char *path) {
    FILE *fp;
    char line[SPRYZEX_MAX_OUTPUT_COLS];

    fp = fopen(path, "r");
    if (fp == NULL) {
        return -1;
    }

    output_buffer_clear(buffer);
    while (fgets(line, sizeof(line), fp) != NULL) {
        output_buffer_append_text(buffer, line);
    }

    fclose(fp);

    if (buffer->line_count == 0) {
        output_buffer_append_line(buffer, "(empty file)");
    }

    return 0;
}

static int load_object_file(OutputBuffer *buffer, const char *path) {
    FILE *fp;
    unsigned int word;
    int count;

    fp = fopen(path, "rb");
    if (fp == NULL) {
        return -1;
    }

    output_buffer_clear(buffer);
    output_buffer_append_line(buffer, "Addr    Word");
    output_buffer_append_line(buffer, "------  --------");

    count = 0;
    while (fread(&word, sizeof(word), 1, fp) == 1) {
        char row[64];
        snprintf(row, sizeof(row), "%06X  %08X", count * 4, word);
        output_buffer_append_line(buffer, row);
        count++;
    }

    if (count == 0) {
        output_buffer_append_line(buffer, "(empty object file)");
    }

    fclose(fp);
    return 0;
}

int runner_load_selected_artifact(AppState *state) {
    int rc;

    output_buffer_clear(&state->artifact_output);

    if (state->output_mode == OUTPUT_MODE_LIVE) {
        return 0;
    }

    if (state->output_mode == OUTPUT_MODE_LOG) {
        rc = load_text_file(&state->artifact_output, state->log_path);
    } else if (state->output_mode == OUTPUT_MODE_LST) {
        rc = load_text_file(&state->artifact_output, state->lst_path);
    } else {
        rc = load_object_file(&state->artifact_output, state->obj_path);
    }

    if (rc != 0) {
        char msg[PATH_MAX + 64];
        const char *path = state->log_path;

        if (state->output_mode == OUTPUT_MODE_LST) {
            path = state->lst_path;
        } else if (state->output_mode == OUTPUT_MODE_OBJ) {
            path = state->obj_path;
        }

        snprintf(msg, sizeof(msg), "Artifact not available: %s", path);
        output_buffer_append_line(&state->artifact_output, msg);
        return -1;
    }

    return 0;
}

void runner_set_output_mode(AppState *state, OutputMode mode) {
    state->output_mode = mode;
    if (mode != OUTPUT_MODE_LIVE) {
        runner_load_selected_artifact(state);
    }
}
