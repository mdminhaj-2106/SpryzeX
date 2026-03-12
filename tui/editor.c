#include "editor.h"

#include <ctype.h>
#include <errno.h>
#include <stdio.h>
#include <string.h>

#include <ncurses.h>

static int line_length(const EditorState *editor, int row) {
    return (int)strlen(editor->lines[row]);
}

static void clamp_cursor(EditorState *editor) {
    int len;

    if (editor->cursor_row < 0) {
        editor->cursor_row = 0;
    }
    if (editor->cursor_row >= editor->line_count) {
        editor->cursor_row = editor->line_count - 1;
    }

    len = line_length(editor, editor->cursor_row);
    if (editor->cursor_col < 0) {
        editor->cursor_col = 0;
    }
    if (editor->cursor_col > len) {
        editor->cursor_col = len;
    }
}

static int insert_char(EditorState *editor, int c) {
    char *line;
    int len;

    line = editor->lines[editor->cursor_row];
    len = (int)strlen(line);

    if (len >= SPRYZEX_MAX_COLS - 1 || editor->cursor_col > len) {
        return 0;
    }

    memmove(&line[editor->cursor_col + 1],
            &line[editor->cursor_col],
            (size_t)(len - editor->cursor_col + 1));
    line[editor->cursor_col] = (char)c;
    editor->cursor_col++;
    editor->dirty = 1;
    return 1;
}

static int insert_newline(EditorState *editor) {
    char *line;
    int len;
    int i;

    if (editor->line_count >= SPRYZEX_MAX_LINES) {
        return 0;
    }

    line = editor->lines[editor->cursor_row];
    len = (int)strlen(line);

    if (editor->cursor_col > len) {
        editor->cursor_col = len;
    }

    for (i = editor->line_count; i > editor->cursor_row + 1; --i) {
        strcpy(editor->lines[i], editor->lines[i - 1]);
    }

    strcpy(editor->lines[editor->cursor_row + 1], &line[editor->cursor_col]);
    line[editor->cursor_col] = '\0';

    editor->line_count++;
    editor->cursor_row++;
    editor->cursor_col = 0;
    editor->dirty = 1;
    return 1;
}

static int backspace(EditorState *editor) {
    char *line;
    int len;

    if (editor->cursor_col > 0) {
        line = editor->lines[editor->cursor_row];
        len = (int)strlen(line);
        memmove(&line[editor->cursor_col - 1],
                &line[editor->cursor_col],
                (size_t)(len - editor->cursor_col + 1));
        editor->cursor_col--;
        editor->dirty = 1;
        return 1;
    }

    if (editor->cursor_row <= 0) {
        return 0;
    }

    {
        int prev_row;
        int prev_len;
        int curr_len;
        int i;

        prev_row = editor->cursor_row - 1;
        prev_len = (int)strlen(editor->lines[prev_row]);
        curr_len = (int)strlen(editor->lines[editor->cursor_row]);

        if (prev_len + curr_len >= SPRYZEX_MAX_COLS) {
            return 0;
        }

        strcat(editor->lines[prev_row], editor->lines[editor->cursor_row]);

        for (i = editor->cursor_row; i < editor->line_count - 1; ++i) {
            strcpy(editor->lines[i], editor->lines[i + 1]);
        }

        editor->line_count--;
        editor->cursor_row = prev_row;
        editor->cursor_col = prev_len;
        editor->dirty = 1;
    }

    return 1;
}

static int delete_char(EditorState *editor) {
    char *line;
    int len;

    line = editor->lines[editor->cursor_row];
    len = (int)strlen(line);

    if (editor->cursor_col < len) {
        memmove(&line[editor->cursor_col],
                &line[editor->cursor_col + 1],
                (size_t)(len - editor->cursor_col));
        editor->dirty = 1;
        return 1;
    }

    if (editor->cursor_row >= editor->line_count - 1) {
        return 0;
    }

    {
        int next_len;
        int i;

        next_len = (int)strlen(editor->lines[editor->cursor_row + 1]);
        if (len + next_len >= SPRYZEX_MAX_COLS) {
            return 0;
        }

        strcat(editor->lines[editor->cursor_row], editor->lines[editor->cursor_row + 1]);
        for (i = editor->cursor_row + 1; i < editor->line_count - 1; ++i) {
            strcpy(editor->lines[i], editor->lines[i + 1]);
        }

        editor->line_count--;
        editor->dirty = 1;
    }

    return 1;
}

static void reset_editor(EditorState *editor) {
    memset(editor, 0, sizeof(*editor));
    editor->line_count = 1;
    editor->lines[0][0] = '\0';
}

int editor_load_file(EditorState *editor, const char *path, char *err, size_t err_size) {
    FILE *fp;
    char buf[SPRYZEX_MAX_COLS * 2];
    int line_count;

    fp = fopen(path, "r");
    if (fp == NULL) {
        if (err != NULL && err_size > 0) {
            snprintf(err, err_size, "Could not open %s: %s", path, strerror(errno));
        }
        reset_editor(editor);
        return -1;
    }

    reset_editor(editor);
    line_count = 0;

    while (fgets(buf, sizeof(buf), fp) != NULL) {
        size_t len;

        if (line_count >= SPRYZEX_MAX_LINES) {
            break;
        }

        len = strlen(buf);
        while (len > 0 && (buf[len - 1] == '\n' || buf[len - 1] == '\r')) {
            buf[len - 1] = '\0';
            len--;
        }

        strncpy(editor->lines[line_count], buf, SPRYZEX_MAX_COLS - 1);
        editor->lines[line_count][SPRYZEX_MAX_COLS - 1] = '\0';
        line_count++;
    }

    fclose(fp);

    if (line_count == 0) {
        line_count = 1;
        editor->lines[0][0] = '\0';
    }

    editor->line_count = line_count;
    editor->cursor_row = 0;
    editor->cursor_col = 0;
    editor->row_offset = 0;
    editor->col_offset = 0;
    editor->dirty = 0;

    if (err != NULL && err_size > 0) {
        err[0] = '\0';
    }

    return 0;
}

int editor_save_file(EditorState *editor, const char *path, char *err, size_t err_size) {
    FILE *fp;
    int i;

    fp = fopen(path, "w");
    if (fp == NULL) {
        if (err != NULL && err_size > 0) {
            snprintf(err, err_size, "Could not save %s: %s", path, strerror(errno));
        }
        return -1;
    }

    for (i = 0; i < editor->line_count; ++i) {
        fputs(editor->lines[i], fp);
        if (i != editor->line_count - 1) {
            fputc('\n', fp);
        }
    }

    fclose(fp);
    editor->dirty = 0;

    if (err != NULL && err_size > 0) {
        err[0] = '\0';
    }

    return 0;
}

int editor_handle_key(EditorState *editor, int key) {
    switch (key) {
        case KEY_LEFT:
            if (editor->cursor_col > 0) {
                editor->cursor_col--;
            } else if (editor->cursor_row > 0) {
                editor->cursor_row--;
                editor->cursor_col = line_length(editor, editor->cursor_row);
            }
            return 1;

        case KEY_RIGHT:
            if (editor->cursor_col < line_length(editor, editor->cursor_row)) {
                editor->cursor_col++;
            } else if (editor->cursor_row < editor->line_count - 1) {
                editor->cursor_row++;
                editor->cursor_col = 0;
            }
            return 1;

        case KEY_UP:
            if (editor->cursor_row > 0) {
                editor->cursor_row--;
                clamp_cursor(editor);
            }
            return 1;

        case KEY_DOWN:
            if (editor->cursor_row < editor->line_count - 1) {
                editor->cursor_row++;
                clamp_cursor(editor);
            }
            return 1;

        case KEY_HOME:
            editor->cursor_col = 0;
            return 1;

        case KEY_END:
            editor->cursor_col = line_length(editor, editor->cursor_row);
            return 1;

        case KEY_NPAGE:
            editor->cursor_row += 10;
            if (editor->cursor_row >= editor->line_count) {
                editor->cursor_row = editor->line_count - 1;
            }
            clamp_cursor(editor);
            return 1;

        case KEY_PPAGE:
            editor->cursor_row -= 10;
            if (editor->cursor_row < 0) {
                editor->cursor_row = 0;
            }
            clamp_cursor(editor);
            return 1;

        case KEY_BACKSPACE:
        case 127:
        case 8:
            return backspace(editor);

        case KEY_DC:
            return delete_char(editor);

        case '\n':
        case '\r':
        case KEY_ENTER:
            return insert_newline(editor);

        case '\t': {
            int i;
            for (i = 0; i < 4; ++i) {
                if (!insert_char(editor, ' ')) {
                    break;
                }
            }
            return 1;
        }

        default:
            if (isprint(key)) {
                return insert_char(editor, key);
            }
            break;
    }

    return 0;
}

void editor_ensure_cursor_visible(EditorState *editor, int view_rows, int view_cols) {
    if (view_rows < 1) {
        view_rows = 1;
    }
    if (view_cols < 1) {
        view_cols = 1;
    }

    if (editor->cursor_row < editor->row_offset) {
        editor->row_offset = editor->cursor_row;
    }
    if (editor->cursor_row >= editor->row_offset + view_rows) {
        editor->row_offset = editor->cursor_row - view_rows + 1;
    }

    if (editor->cursor_col < editor->col_offset) {
        editor->col_offset = editor->cursor_col;
    }
    if (editor->cursor_col >= editor->col_offset + view_cols) {
        editor->col_offset = editor->cursor_col - view_cols + 1;
    }

    if (editor->row_offset < 0) {
        editor->row_offset = 0;
    }
    if (editor->col_offset < 0) {
        editor->col_offset = 0;
    }
}
