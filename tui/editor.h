#ifndef SPRYZEX_EDITOR_H
#define SPRYZEX_EDITOR_H

#include <stddef.h>

#include "state.h"

int editor_load_file(EditorState *editor, const char *path, char *err, size_t err_size);
int editor_save_file(EditorState *editor, const char *path, char *err, size_t err_size);

/* Returns 1 if key was handled by editor, else 0. */
int editor_handle_key(EditorState *editor, int key);

void editor_ensure_cursor_visible(EditorState *editor, int view_rows, int view_cols);

#endif
