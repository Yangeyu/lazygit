#!/usr/bin/env bash
#
# setup-ai-commit.sh
#
# Sets up the "generate commit message" feature of this lazygit fork on a new
# machine:
#   1. installs the lazygit binary from this repo (go install)
#   2. adds an aichat "qwen" client pointing at the DashScope endpoint
#   3. adds the generateCommitMessageCommand to the lazygit config
#
# It is idempotent and merges into existing config files (your other aichat
# clients / lazygit settings are preserved). No API key is written anywhere:
# the command reads QWEN_API_KEY from your existing DASHSCOPE_API_KEY at runtime.
#
# Requirements: bash, python3, and (for step 1) go. aichat must be installed
# separately: `brew install aichat`.

set -euo pipefail

API_BASE="https://llm-r1no6xbr99vaoqm4.cn-beijing.maas.aliyuncs.com/compatible-mode/v1"
MODEL="qwen:qwen3.7-plus" # switch to qwen:qwen3.7-max for higher quality
GEN_PROMPT="Generate a Conventional Commits message for the staged diff provided on stdin. The summary line must be <type>(<scope>): <description>, where type is one of feat, fix, docs, style, refactor, perf, test, build, ci, chore (scope is optional). Use the imperative mood, a lowercase description, and no trailing period. Add a blank line and a body of - bullet points only if it conveys useful detail. Output ONLY the commit message: no code fences, no preamble, no quotes."
GEN_COMMAND="QWEN_API_KEY=\"\$DASHSCOPE_API_KEY\" aichat -m ${MODEL} -S \"${GEN_PROMPT}\""

note() { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
warn() { printf '\033[1;33mwarning:\033[0m %s\n' "$1"; }

# --- 0. sanity checks -------------------------------------------------------
command -v python3 >/dev/null || { echo "python3 is required"; exit 1; }

if [ -z "${DASHSCOPE_API_KEY:-}" ]; then
  warn "DASHSCOPE_API_KEY is not set in this shell. The feature needs it at runtime."
  warn "Add 'export DASHSCOPE_API_KEY=sk-...' to your ~/.zshrc, then restart the shell."
fi

command -v aichat >/dev/null || warn "aichat not found on PATH — install it with 'brew install aichat'."

# --- 1. install the lazygit binary ------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if command -v go >/dev/null && grep -q "module github.com/jesseduffield/lazygit" "$SCRIPT_DIR/go.mod" 2>/dev/null; then
  note "Installing lazygit from $SCRIPT_DIR (go install)"
  (cd "$SCRIPT_DIR" && go install .)
  note "Installed to $(go env GOPATH)/bin/lazygit — make sure that dir is early in your PATH."
else
  warn "Skipping build (go not found or not run from the repo). Build manually with 'go install .'."
fi

# --- 2. resolve config paths ------------------------------------------------
if command -v aichat >/dev/null; then
  AICHAT_CONFIG="$(aichat --info 2>/dev/null | awk '/^config_file/{print $2}')"
fi
if [ -z "${AICHAT_CONFIG:-}" ]; then
  case "$(uname -s)" in
    Darwin) AICHAT_CONFIG="$HOME/Library/Application Support/aichat/config.yaml" ;;
    *)      AICHAT_CONFIG="${XDG_CONFIG_HOME:-$HOME/.config}/aichat/config.yaml" ;;
  esac
fi

case "$(uname -s)" in
  Darwin) LG_CONFIG_DIR="${XDG_CONFIG_HOME:+$XDG_CONFIG_HOME/lazygit}"; LG_CONFIG_DIR="${LG_CONFIG_DIR:-$HOME/Library/Application Support/lazygit}" ;;
  *)      LG_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/lazygit" ;;
esac
LG_CONFIG="$LG_CONFIG_DIR/config.yml"

# --- 3. merge the aichat qwen client ----------------------------------------
note "Updating aichat config: $AICHAT_CONFIG"
API_BASE="$API_BASE" python3 - "$AICHAT_CONFIG" <<'PY'
import os, sys, yaml
path = sys.argv[1]
cfg = {}
if os.path.exists(path):
    cfg = yaml.safe_load(open(path)) or {}
cfg.setdefault("clients", [])
clients = cfg["clients"]
qwen = {
    "type": "openai-compatible",
    "name": "qwen",
    "api_base": os.environ["API_BASE"],
    "models": [
        {"name": "qwen3.7-plus", "max_input_tokens": 32768},
        {"name": "qwen3.7-max", "max_input_tokens": 32768},
    ],
    "patch": {"chat_completions": {r"qwen3\.7.*": {"body": {"enable_thinking": False}}}},
}
for i, c in enumerate(clients):
    if isinstance(c, dict) and c.get("name") == "qwen":
        clients[i] = qwen
        break
else:
    clients.append(qwen)
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w") as f:
    yaml.safe_dump(cfg, f, sort_keys=False, allow_unicode=True, default_flow_style=False)
print("  qwen client written")
PY

# --- 4. merge the lazygit command -------------------------------------------
note "Updating lazygit config: $LG_CONFIG"
GEN_COMMAND="$GEN_COMMAND" python3 - "$LG_CONFIG" <<'PY'
import os, sys, yaml
path = sys.argv[1]
cfg = {}
if os.path.exists(path):
    cfg = yaml.safe_load(open(path)) or {}
commit = cfg.setdefault("git", {}).setdefault("commit", {})
commit["generateCommitMessageCommand"] = os.environ["GEN_COMMAND"]
commit.setdefault("autoGenerateCommitMessage", False)
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w") as f:
    yaml.safe_dump(cfg, f, sort_keys=False, allow_unicode=True, default_flow_style=False)
print("  generateCommitMessageCommand written")
PY

note "Done. Open lazygit, stage files, press 'c', then <ctrl+g> to generate, <enter> to accept."
