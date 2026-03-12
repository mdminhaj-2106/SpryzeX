#include "mascot.h"

#include <stdio.h>
#include <string.h>

static const char *MASCOT_FRAMES[][9] = {
    {
        "        .-''''''-.        ",
        "      .'  .--.    '.      ",
        "     /   /_  _\\     \\     ",
        "    |    (o)(o)      |    ",
        "    |    .-__-.      |    ",
        "    |   /|____|\\     |    ",
        "     \\    \\__/     _/     ",
        "      '._      _.-'       ",
        "         '--..--'         "
    },
    {
        "        .-''''''-.        ",
        "      .'  .--.    '.      ",
        "     /   /_  _\\     \\     ",
        "    |    (o)(-)      |    ",
        "    |    .-__-.    __|    ",
        "    |   /|____|\\  / /    ",
        "     \\    \\__/   /_/      ",
        "      '._      _.-'       ",
        "         '--..--'         "
    },
    {
        "        .-''''''-.        ",
        "      .'  .--.    '.      ",
        "     /   /_  _\\     \\     ",
        "    |    [==][==]    |    ",
        "    |    .-__-.      |    ",
        "    |   /|____|\\     |    ",
        "     \\    \\__/     _/     ",
        "      '._      _.-'       ",
        "         '--..--'         "
    },
    {
        "        .-''''''-.        ",
        "      .'  .--.    '.      ",
        "     /   /_  _\\     \\     ",
        "    |    (-)(o)      |    ",
        "    |    .-__-.      |    ",
        "    |   /|____|\\   _|    ",
        "     \\    \\__/   (_/     ",
        "      '._      _.-'       ",
        "         '--..--'         "
    }
};

static const char *SPARK_LINES[] = {
    "  .   *    .      *    .    ",
    "    *    .    *      .      ",
    " .    .      *   .      *   ",
    "    .      *    .    *      "
};

static const char *TAG_LINES[] = {
    "ASSEMBLE FAST",
    "EMULATE WITH CLARITY",
    "TRACE. FIX. SHIP.",
    "SPRYBOT ONLINE"
};

static const char *PULSE_LINES[] = {
    "core [==      ]",
    "core [====    ]",
    "core [======  ]",
    "core [========]"
};

static const int FRAME_COUNT = (int)(sizeof(MASCOT_FRAMES) / sizeof(MASCOT_FRAMES[0]));

void mascot_tick(AppState *state, long long now_ms) {
    if (state->last_mascot_tick_ms == 0) {
        state->last_mascot_tick_ms = now_ms;
        return;
    }

    if (now_ms - state->last_mascot_tick_ms >= 160) {
        state->mascot_frame = (state->mascot_frame + 1) % FRAME_COUNT;
        state->last_mascot_tick_ms = now_ms;
    }
}

void mascot_render(WINDOW *win, const AppState *state) {
    int rows;
    int cols;
    int i;
    int start_row;
    int start_col;
    int frame_idx;
    int bob;
    const char **frame;

    getmaxyx(win, rows, cols);

    frame_idx = state->mascot_frame % FRAME_COUNT;
    frame = MASCOT_FRAMES[frame_idx];

    for (i = 1; i < rows - 1; ++i) {
        mvwhline(win, i, 1, ' ', cols - 2);
    }

    wattron(win, A_BOLD);
    mvwprintw(win, 1, 2, "%s", SPARK_LINES[frame_idx]);
    wattroff(win, A_BOLD);

    bob = 0;
    if (frame_idx == 1) {
        bob = 1;
    } else if (frame_idx == 3) {
        bob = -1;
    }

    start_row = (rows - 9) / 2 + bob;
    if (start_row < 3) {
        start_row = 3;
    }

    for (i = 0; i < 9 && (start_row + i) < rows - 2; ++i) {
        int frame_len = (int)strlen(frame[i]);
        start_col = (cols - frame_len) / 2;
        if (start_col < 1) {
            start_col = 1;
        }

        if (i >= 3 && i <= 6) {
            wattron(win, A_BOLD);
            mvwprintw(win, start_row + i, start_col, "%s", frame[i]);
            wattroff(win, A_BOLD);
        } else {
            mvwprintw(win, start_row + i, start_col, "%s", frame[i]);
        }
    }

    if (rows >= 6) {
        char banner[64];
        snprintf(banner, sizeof(banner), "[%s]", TAG_LINES[frame_idx]);
        start_col = (cols - (int)strlen(banner)) / 2;
        if (start_col < 1) {
            start_col = 1;
        }
        mvwprintw(win, rows - 4, start_col, "%s", banner);
    }

    mvwprintw(win, rows - 2, 2, "%s", PULSE_LINES[frame_idx]);
}
