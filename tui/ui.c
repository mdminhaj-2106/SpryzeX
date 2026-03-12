#include "ui.h"

#include <ctype.h>
#include <stdio.h>
#include <string.h>

#include "editor.h"
#include "mascot.h"

#define COLOR_TITLE 1
#define COLOR_STATUS 2
#define COLOR_ACCENT 3
#define COLOR_BORDER 4
#define COLOR_GUTTER 5
#define COLOR_CURSORLINE 6
#define COLOR_MASCOT 7
#define COLOR_ERROR 8
#define COLOR_WARNING 9
#define COLOR_MUTED 10

static int contains_ci(const char *text, const char *needle) {
    size_t i;
    size_t nlen;

    if (text == NULL || needle == NULL) {
        return 0;
    }

    nlen = strlen(needle);
    if (nlen == 0) {
        return 1;
    }

    for (i = 0; text[i] != '\0'; ++i) {
        size_t j = 0;

        while (needle[j] != '\0' && text[i + j] != '\0' &&
               tolower((unsigned char)text[i + j]) == tolower((unsigned char)needle[j])) {
            j++;
        }

        if (j == nlen) {
            return 1;
        }
    }

    return 0;
}

static void draw_panel_frame(WINDOW *win, const char *title) {
    wattron(win, COLOR_PAIR(COLOR_BORDER));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(COLOR_BORDER));

    if (title != NULL && title[0] != '\0') {
        wattron(win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
        mvwprintw(win, 0, 2, " %s ", title);
        wattroff(win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    }
}

static void destroy_menu(UIContext *ui) {
    int i;

    if (ui->mode_menu != NULL) {
        unpost_menu(ui->mode_menu);
    }

    if (ui->mode_subwin != NULL) {
        delwin(ui->mode_subwin);
        ui->mode_subwin = NULL;
    }

    if (ui->mode_menu != NULL) {
        free_menu(ui->mode_menu);
        ui->mode_menu = NULL;
    }

    for (i = 0; i < 4; ++i) {
        if (ui->mode_items[i] != NULL) {
            free_item(ui->mode_items[i]);
            ui->mode_items[i] = NULL;
        }
    }
}

static void destroy_windows(UIContext *ui) {
    if (ui->title_panel != NULL) {
        del_panel(ui->title_panel);
        ui->title_panel = NULL;
    }
    if (ui->mascot_panel != NULL) {
        del_panel(ui->mascot_panel);
        ui->mascot_panel = NULL;
    }
    if (ui->editor_panel != NULL) {
        del_panel(ui->editor_panel);
        ui->editor_panel = NULL;
    }
    if (ui->output_panel != NULL) {
        del_panel(ui->output_panel);
        ui->output_panel = NULL;
    }
    if (ui->status_panel != NULL) {
        del_panel(ui->status_panel);
        ui->status_panel = NULL;
    }

    if (ui->title_win != NULL) {
        delwin(ui->title_win);
        ui->title_win = NULL;
    }
    if (ui->mascot_win != NULL) {
        delwin(ui->mascot_win);
        ui->mascot_win = NULL;
    }
    if (ui->editor_win != NULL) {
        delwin(ui->editor_win);
        ui->editor_win = NULL;
    }
    if (ui->output_win != NULL) {
        delwin(ui->output_win);
        ui->output_win = NULL;
    }
    if (ui->status_win != NULL) {
        delwin(ui->status_win);
        ui->status_win = NULL;
    }
}

static int create_windows(UIContext *ui) {
    int title_h;
    int status_h;
    int output_h;
    int top_h;
    int mascot_w;
    int editor_w;

    getmaxyx(stdscr, ui->screen_rows, ui->screen_cols);

    if (ui->screen_rows < 20 || ui->screen_cols < 70) {
        return -1;
    }

    title_h = 3;
    status_h = 3;

    output_h = ui->screen_rows / 3;
    if (output_h < 7) {
        output_h = 7;
    }

    top_h = ui->screen_rows - title_h - output_h - status_h;
    if (top_h < 8) {
        top_h = 8;
        output_h = ui->screen_rows - title_h - top_h - status_h;
    }

    if (output_h < 6) {
        return -1;
    }

    mascot_w = ui->screen_cols / 3;
    if (mascot_w < 22) {
        mascot_w = 22;
    }
    if (mascot_w > 40) {
        mascot_w = 40;
    }

    editor_w = ui->screen_cols - mascot_w;
    if (editor_w < 34) {
        editor_w = 34;
        mascot_w = ui->screen_cols - editor_w;
    }

    if (mascot_w < 18 || editor_w < 30) {
        return -1;
    }

    ui->title_win = newwin(title_h, ui->screen_cols, 0, 0);
    ui->mascot_win = newwin(top_h, mascot_w, title_h, 0);
    ui->editor_win = newwin(top_h, editor_w, title_h, mascot_w);
    ui->output_win = newwin(output_h, ui->screen_cols, title_h + top_h, 0);
    ui->status_win = newwin(status_h, ui->screen_cols, title_h + top_h + output_h, 0);

    if (ui->title_win == NULL || ui->mascot_win == NULL || ui->editor_win == NULL ||
        ui->output_win == NULL || ui->status_win == NULL) {
        return -1;
    }

    keypad(ui->editor_win, TRUE);

    ui->title_panel = new_panel(ui->title_win);
    ui->mascot_panel = new_panel(ui->mascot_win);
    ui->editor_panel = new_panel(ui->editor_win);
    ui->output_panel = new_panel(ui->output_win);
    ui->status_panel = new_panel(ui->status_win);

    if (ui->title_panel == NULL || ui->mascot_panel == NULL || ui->editor_panel == NULL ||
        ui->output_panel == NULL || ui->status_panel == NULL) {
        return -1;
    }

    ui->editor_view_rows = top_h - 2;
    ui->editor_view_cols = editor_w - 2 - 8;
    if (ui->editor_view_cols < 12) {
        ui->editor_view_cols = 12;
    }

    ui->output_view_rows = output_h - 3;
    ui->output_view_cols = ui->screen_cols - 4;
    if (ui->output_view_rows < 1) {
        ui->output_view_rows = 1;
    }

    return 0;
}

static int create_menu(UIContext *ui) {
    int tab_w;
    int tab_x;

    ui->mode_items[0] = new_item(" LIVE ", "");
    ui->mode_items[1] = new_item(" LOG ", "");
    ui->mode_items[2] = new_item(" LST ", "");
    ui->mode_items[3] = new_item(" OBJ ", "");
    ui->mode_items[4] = NULL;

    if (ui->mode_items[0] == NULL || ui->mode_items[1] == NULL || ui->mode_items[2] == NULL ||
        ui->mode_items[3] == NULL) {
        return -1;
    }

    ui->mode_menu = new_menu(ui->mode_items);
    if (ui->mode_menu == NULL) {
        return -1;
    }

    set_menu_mark(ui->mode_menu, "");
    set_menu_format(ui->mode_menu, 1, 4);
    set_menu_fore(ui->mode_menu, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    set_menu_back(ui->mode_menu, COLOR_PAIR(COLOR_MUTED));
    set_menu_grey(ui->mode_menu, A_DIM);
    set_menu_spacing(ui->mode_menu, 0, 1, 1);

    set_menu_win(ui->mode_menu, ui->output_win);

    tab_w = 30;
    tab_x = ui->screen_cols - tab_w - 3;
    if (tab_x < 2) {
        tab_x = 2;
    }

    ui->mode_subwin = derwin(ui->output_win, 1, tab_w, 1, tab_x);
    if (ui->mode_subwin == NULL) {
        return -1;
    }
    set_menu_sub(ui->mode_menu, ui->mode_subwin);

    if (post_menu(ui->mode_menu) != E_OK) {
        return -1;
    }

    return 0;
}

static void draw_title(const UIContext *ui, const AppState *state) {
    const char *base;
    char mode_buf[32];

    base = strrchr(state->current_file, '/');
    if (base == NULL) {
        base = state->current_file;
    } else {
        base++;
    }

    werase(ui->title_win);
    wbkgd(ui->title_win, COLOR_PAIR(COLOR_TITLE));
    wattron(ui->title_win, COLOR_PAIR(COLOR_BORDER));
    box(ui->title_win, 0, 0);
    wattroff(ui->title_win, COLOR_PAIR(COLOR_BORDER));

    wattron(ui->title_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    mvwprintw(ui->title_win, 1, 2, "SPRYZEX  //  TERMINAL STUDIO");
    wattroff(ui->title_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);

    wattron(ui->title_win, COLOR_PAIR(COLOR_MUTED));
    mvwprintw(ui->title_win, 1, 34, "File: %s%s", base, state->editor.dirty ? " *" : "");
    wattroff(ui->title_win, COLOR_PAIR(COLOR_MUTED));

    snprintf(mode_buf, sizeof(mode_buf), "[%s]", output_mode_name(state->output_mode));
    wattron(ui->title_win, COLOR_PAIR(COLOR_GUTTER) | A_BOLD);
    mvwprintw(ui->title_win, 1, ui->screen_cols - (int)strlen(mode_buf) - 3, "%s", mode_buf);
    wattroff(ui->title_win, COLOR_PAIR(COLOR_GUTTER) | A_BOLD);
}

static void draw_editor(const UIContext *ui, AppState *state) {
    int h;
    int w;
    int row;
    int y;
    int max_row;

    getmaxyx(ui->editor_win, h, w);

    werase(ui->editor_win);
    draw_panel_frame(ui->editor_win, "CODE");

    editor_ensure_cursor_visible(&state->editor, ui->editor_view_rows, ui->editor_view_cols);

    max_row = state->editor.row_offset + ui->editor_view_rows;
    if (max_row > state->editor.line_count) {
        max_row = state->editor.line_count;
    }

    y = 1;
    for (row = state->editor.row_offset; row < max_row; ++row) {
        char visible[SPRYZEX_MAX_COLS];
        int len;
        int is_current;

        len = (int)strlen(state->editor.lines[row]);
        is_current = (row == state->editor.cursor_row);

        if (state->editor.col_offset >= len) {
            visible[0] = '\0';
        } else {
            strncpy(visible,
                    state->editor.lines[row] + state->editor.col_offset,
                    ui->editor_view_cols);
            visible[ui->editor_view_cols] = '\0';
        }

        wattron(ui->editor_win, COLOR_PAIR(COLOR_GUTTER) | A_BOLD);
        mvwprintw(ui->editor_win, y, 1, "%6d", row + 1);
        wattroff(ui->editor_win, COLOR_PAIR(COLOR_GUTTER) | A_BOLD);
        wattron(ui->editor_win, COLOR_PAIR(COLOR_BORDER));
        mvwaddch(ui->editor_win, y, 7, ACS_VLINE);
        wattroff(ui->editor_win, COLOR_PAIR(COLOR_BORDER));

        if (is_current) {
            wattron(ui->editor_win, COLOR_PAIR(COLOR_CURSORLINE));
        }

        mvwprintw(ui->editor_win, y, 8, "%-*s", ui->editor_view_cols, visible);

        if (is_current) {
            wattroff(ui->editor_win, COLOR_PAIR(COLOR_CURSORLINE));
        }

        y++;
    }

    for (; y < h - 1; ++y) {
        wattron(ui->editor_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
        mvwprintw(ui->editor_win, y, 1, "%6s", "~");
        wattroff(ui->editor_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
        wattron(ui->editor_win, COLOR_PAIR(COLOR_BORDER));
        mvwaddch(ui->editor_win, y, 7, ACS_VLINE);
        wattroff(ui->editor_win, COLOR_PAIR(COLOR_BORDER));
        mvwprintw(ui->editor_win, y, 8, "%-*s", ui->editor_view_cols, "");
    }

    wattron(ui->editor_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
    mvwprintw(ui->editor_win,
              h - 1,
              w - 18,
              "Ln %d  Col %d",
              state->editor.cursor_row + 1,
              state->editor.cursor_col + 1);
    wattroff(ui->editor_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);

    {
        int cursor_screen_y = 1 + (state->editor.cursor_row - state->editor.row_offset);
        int cursor_screen_x = 8 + (state->editor.cursor_col - state->editor.col_offset);

        if (cursor_screen_y >= 1 && cursor_screen_y < h - 1 &&
            cursor_screen_x >= 8 && cursor_screen_x < w - 1) {
            wmove(ui->editor_win, cursor_screen_y, cursor_screen_x);
        }
    }
}

static const OutputBuffer *active_output_buffer(const AppState *state) {
    if (state->output_mode == OUTPUT_MODE_LIVE) {
        return &state->live_output;
    }
    return &state->artifact_output;
}

static void draw_output(const UIContext *ui, const AppState *state) {
    const OutputBuffer *buffer;
    int i;
    int start;
    int end;
    char stats[48];

    buffer = active_output_buffer(state);

    werase(ui->output_win);
    draw_panel_frame(ui->output_win, "CONSOLE");

    wattron(ui->output_win, COLOR_PAIR(COLOR_MUTED));
    mvwprintw(ui->output_win, 1, 2, "View");
    wattroff(ui->output_win, COLOR_PAIR(COLOR_MUTED));

    snprintf(stats,
             sizeof(stats),
             "lines:%d  scroll:%d",
             buffer->line_count,
             buffer->scroll);
    wattron(ui->output_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
    mvwprintw(ui->output_win, 1, 14, "%s", stats);
    wattroff(ui->output_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);

    if (ui->mode_menu != NULL && state->output_mode >= OUTPUT_MODE_LIVE &&
        state->output_mode <= OUTPUT_MODE_OBJ) {
        set_current_item(ui->mode_menu, ui->mode_items[state->output_mode]);
    }

    start = buffer->line_count - ui->output_view_rows - buffer->scroll;
    if (start < 0) {
        start = 0;
    }

    end = start + ui->output_view_rows;
    if (end > buffer->line_count) {
        end = buffer->line_count;
    }

    if (buffer->line_count == 0) {
        wattron(ui->output_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
        mvwprintw(ui->output_win, 2, 2, "No output yet. Build with B or run with R.");
        wattroff(ui->output_win, COLOR_PAIR(COLOR_MUTED) | A_DIM);
    }

    for (i = start; i < end; ++i) {
        int attr = COLOR_PAIR(COLOR_MUTED);

        if (contains_ci(buffer->lines[i], "error")) {
            attr = COLOR_PAIR(COLOR_ERROR) | A_BOLD;
        } else if (contains_ci(buffer->lines[i], "warning")) {
            attr = COLOR_PAIR(COLOR_WARNING) | A_BOLD;
        } else if (contains_ci(buffer->lines[i], "halt") || contains_ci(buffer->lines[i], "success")) {
            attr = COLOR_PAIR(COLOR_ACCENT);
        }

        wattron(ui->output_win, attr);
        mvwprintw(ui->output_win,
                  2 + (i - start),
                  2,
                  "%-*.*s",
                  ui->output_view_cols,
                  ui->output_view_cols,
                  buffer->lines[i]);
        wattroff(ui->output_win, attr);
    }
}

static void draw_status(const UIContext *ui, const AppState *state) {
    char help[256];
    int status_x;

    werase(ui->status_win);
    wbkgd(ui->status_win, COLOR_PAIR(COLOR_STATUS));

    wattron(ui->status_win, COLOR_PAIR(COLOR_BORDER));
    box(ui->status_win, 0, 0);
    wattroff(ui->status_win, COLOR_PAIR(COLOR_BORDER));

    snprintf(help,
             sizeof(help),
             "B Build  R Run  S Save  O Open  Tab/F2-F5 Views  [ ] Scroll  Q Quit");
    wattron(ui->status_win, COLOR_PAIR(COLOR_MUTED) | A_BOLD);
    mvwprintw(ui->status_win, 1, 2, "%.*s", ui->screen_cols - 6, help);
    wattroff(ui->status_win, COLOR_PAIR(COLOR_MUTED) | A_BOLD);

    status_x = ui->screen_cols - (int)strlen(state->status) - 4;
    if (status_x < 2) {
        status_x = 2;
    }

    wattron(ui->status_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    mvwprintw(ui->status_win, 1, status_x, " %s ", state->status);
    wattroff(ui->status_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
}

int ui_init(UIContext *ui) {
    memset(ui, 0, sizeof(*ui));

    initscr();
    cbreak();
    noecho();
    keypad(stdscr, TRUE);
    timeout(50);

    curs_set(1);

    if (has_colors()) {
        start_color();
        use_default_colors();

        init_pair(COLOR_TITLE, COLOR_CYAN, -1);
        init_pair(COLOR_STATUS, COLOR_WHITE, -1);
        init_pair(COLOR_ACCENT, COLOR_YELLOW, -1);
        init_pair(COLOR_BORDER, COLOR_BLUE, -1);
        init_pair(COLOR_GUTTER, COLOR_CYAN, -1);
        init_pair(COLOR_CURSORLINE, COLOR_BLACK, COLOR_CYAN);
        init_pair(COLOR_MASCOT, COLOR_GREEN, -1);
        init_pair(COLOR_ERROR, COLOR_RED, -1);
        init_pair(COLOR_WARNING, COLOR_YELLOW, -1);
        init_pair(COLOR_MUTED, COLOR_WHITE, -1);
    }

    if (create_windows(ui) != 0) {
        return -1;
    }

    if (create_menu(ui) != 0) {
        return -1;
    }

    return 0;
}

void ui_destroy(UIContext *ui) {
    destroy_menu(ui);
    destroy_windows(ui);
    endwin();
}

void ui_handle_resize(UIContext *ui) {
    endwin();
    refresh();
    clear();

    destroy_menu(ui);
    destroy_windows(ui);

    if (create_windows(ui) != 0) {
        return;
    }

    if (create_menu(ui) != 0) {
        destroy_windows(ui);
    }
}

void ui_render(UIContext *ui, AppState *state) {
    if (ui->title_win == NULL || ui->editor_win == NULL || ui->mascot_win == NULL ||
        ui->output_win == NULL || ui->status_win == NULL) {
        erase();
        mvprintw(0, 0, "Terminal too small for SpryzeX UI. Resize (min ~70x20) and continue.");
        refresh();
        return;
    }

    draw_title(ui, state);

    werase(ui->mascot_win);
    draw_panel_frame(ui->mascot_win, "BOT");
    wattron(ui->mascot_win, COLOR_PAIR(COLOR_MASCOT));
    mascot_render(ui->mascot_win, state);
    wattroff(ui->mascot_win, COLOR_PAIR(COLOR_MASCOT));

    draw_editor(ui, state);
    draw_output(ui, state);
    draw_status(ui, state);

    update_panels();
    doupdate();
}

int ui_prompt_input(UIContext *ui, const char *prompt, char *buf, int buf_size) {
    if (ui->status_win == NULL || buf == NULL || buf_size <= 1) {
        return -1;
    }

    timeout(-1);

    werase(ui->status_win);
    wbkgd(ui->status_win, COLOR_PAIR(COLOR_STATUS));
    wattron(ui->status_win, COLOR_PAIR(COLOR_BORDER));
    box(ui->status_win, 0, 0);
    wattroff(ui->status_win, COLOR_PAIR(COLOR_BORDER));

    wattron(ui->status_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    mvwprintw(ui->status_win, 1, 2, "%s", prompt);
    wattroff(ui->status_win, COLOR_PAIR(COLOR_ACCENT) | A_BOLD);
    wrefresh(ui->status_win);

    echo();
    curs_set(1);
    wmove(ui->status_win, 1, (int)strlen(prompt) + 3);
    wgetnstr(ui->status_win, buf, buf_size - 1);
    noecho();

    timeout(50);

    if (buf[0] == '\0') {
        return -1;
    }
    return 0;
}
