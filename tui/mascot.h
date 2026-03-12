#ifndef SPRYZEX_MASCOT_H
#define SPRYZEX_MASCOT_H

#include <ncurses.h>

#include "state.h"

void mascot_tick(AppState *state, long long now_ms);
void mascot_render(WINDOW *win, const AppState *state);

#endif
