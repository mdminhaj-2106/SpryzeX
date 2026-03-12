#ifndef SPRYZEX_RUNNER_H
#define SPRYZEX_RUNNER_H

#include "state.h"

void runner_update_artifact_paths(AppState *state);
int runner_build(AppState *state);
int runner_run_emulator(AppState *state);
int runner_load_selected_artifact(AppState *state);
void runner_set_output_mode(AppState *state, OutputMode mode);

#endif
