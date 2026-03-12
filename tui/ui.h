#ifndef SPRYZEX_UI_H
#define SPRYZEX_UI_H

#include <menu.h>
#include <ncurses.h>
#include <panel.h>

#include "state.h"

typedef struct UIContext {
    WINDOW *title_win;
    WINDOW *mascot_win;
    WINDOW *editor_win;
    WINDOW *output_win;
    WINDOW *status_win;

    PANEL *title_panel;
    PANEL *mascot_panel;
    PANEL *editor_panel;
    PANEL *output_panel;
    PANEL *status_panel;

    MENU *mode_menu;
    ITEM *mode_items[5];
    WINDOW *mode_subwin;

    int screen_rows;
    int screen_cols;

    int editor_view_rows;
    int editor_view_cols;
    int output_view_rows;
    int output_view_cols;
} UIContext;

int ui_init(UIContext *ui);
void ui_destroy(UIContext *ui);
void ui_handle_resize(UIContext *ui);
void ui_render(UIContext *ui, AppState *state);
int ui_prompt_input(UIContext *ui, const char *prompt, char *buf, int buf_size);

#endif
