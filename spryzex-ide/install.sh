#!/usr/bin/env bash
# install.sh — Build and install SPRYZEX IDE
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPRYZEX_DIR="$SCRIPT_DIR"
INSTALL_BIN="/usr/local/bin/spryzex-ide"

echo ""
echo "  ◈ SPRYZEX IDE — Installing"
echo "  ─────────────────────────"
echo ""

# 1. Check Go
if ! command -v go &>/dev/null; then
  echo "  [!] Go is not installed."
  echo "      Install from: https://go.dev/dl/"
  exit 1
fi

GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
echo "  [✓] Go $GO_VER found"

# 2. Build SPRYZEX IDE
echo "  [*] Building spryzex-ide..."
cd "$SPRYZEX_DIR"
GONOSUMDB="*" GOPROXY="direct" GOFLAGS="-mod=mod" go build -o spryzex-ide . 2>&1 | grep -v "^$" || true

if [ ! -f spryzex-ide ]; then
  echo "  [!] Build failed"
  exit 1
fi
echo "  [✓] Built successfully"

# 3. Build C assembler/emulator (optional)
ASM_ROOT="$(dirname "$SPRYZEX_DIR")"
if [ -d "$ASM_ROOT" ] && [ -f "$ASM_ROOT/Makefile" ]; then
  echo "  [*] Building C assembler/emulator..."
  if make -C "$ASM_ROOT" all 2>&1; then
    echo "  [✓] C assembler/emulator built in $ASM_ROOT"
  else
    echo "  [!] C assembler/emulator build failed (optional — spryzex-ide still works)"
  fi
fi

# 4. Install binary
echo "  [*] Installing to $INSTALL_BIN..."
if [ -w "/usr/local/bin" ] || sudo true 2>/dev/null; then
  cp "$SPRYZEX_DIR/spryzex-ide" "$INSTALL_BIN" 2>/dev/null || sudo cp "$SPRYZEX_DIR/spryzex-ide" "$INSTALL_BIN"
  echo "  [✓] Installed to $INSTALL_BIN"
else
  # Install to ~/.local/bin
  mkdir -p "$HOME/.local/bin"
  cp "$SPRYZEX_DIR/spryzex-ide" "$HOME/.local/bin/spryzex-ide"
  echo "  [✓] Installed to $HOME/.local/bin/spryzex-ide"
  echo "      Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo ""
echo "  ─────────────────────────────────────────────"
echo "  SPRYZEX IDE installed!"
echo ""
echo "  Usage:"
echo "    spryzex-ide                   # open SPRYZEX IDE (new file)"
echo "    spryzex-ide file.asm          # open specific file"
echo ""
echo "  KEY BINDINGS (inside editor):"
echo "    i            — Enter INSERT mode"
echo "    ESC          — Return to NORMAL mode"
echo "    :w           — Save"
echo "    :q           — Quit"
echo "    :build       — Assemble"
echo "    :run         — Run emulator"
echo "    Ctrl+B       — Build"
echo "    Ctrl+R       — Run"
echo "    Ctrl+W       — Cycle panel focus"
echo "    Ctrl+]       — Next output tab"
echo "    /pattern     — Search"
echo "    n/N          — Search next/prev"
echo "    gg/G         — Go to top/bottom"
echo "  ─────────────────────────────────────────────"
echo ""
