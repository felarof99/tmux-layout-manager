<div align="center">

# 🪟 layouts

**Tmux layout manager — predefined pane arrangements from a single config.**

*One command. Windows, panes, splits, and commands.*

</div>

Define tmux layouts once in YAML, apply them to any session. Like tmuxinator's layout system, but standalone and minimal — just the layout part, nothing else.

- **Declarative layouts** — windows, panes, split directions, sizes, and startup commands in YAML
- **Apply anywhere** — works on any tmux session, not tied to a specific project
- **fzf picker** — pick a layout interactively when you don't specify one
- **Session creation** — create new tmux sessions with layouts pre-applied
- **Zero state** — no database, no state files, just config

---

## Install

Requires Go 1.21+ and [fzf](https://github.com/junegunn/fzf).

```sh
git clone <repo-url> layouts
cd layouts
make install    # builds and copies to $GOPATH/bin
```

**Fish shell helper** — install the `ly` shorthand function:

```sh
make fish       # copies ly.fish to ~/.config/fish/functions/
```

## Quick Start

```sh
# 1. Create config with example layouts
layouts init

# 2. See what's available
layouts list

# 3. Apply a layout to your current tmux session
layouts apply dev
```

## Config

Location: `~/.config/layouts/config.yaml` (created by `layouts init`)

```yaml
# default: dev
# editor: nvim

layouts:
  dev:
    windows:
      - name: claude
        split: horizontal
        panes:
          - name: claude
            size: "25%"
            cmd: claude --dangerously-skip-permissions
          - name: editor
            size: "50%"
            cmd: nvim .
          - name: shell
            size: "25%"
      - name: test
        panes:
          - name: test-1
            size: "50%"
          - name: test-2
            size: "50%"

  simple:
    windows:
      - name: main
        panes:
          - name: editor
            size: "70%"
            cmd: nvim .
          - name: shell
```

Each layout has one or more **windows**, each with **panes**:

| Field | Description |
|-------|-------------|
| `name` | Window or pane name |
| `split` | `horizontal` (side by side, default) or `vertical` (stacked) |
| `size` | Percentage of the window (e.g. `70%`). Unspecified panes split remaining space equally |
| `cmd` | Command to run in the pane. Empty = shell prompt |

Top-level optional fields:

| Field | Description |
|-------|-------------|
| `default` | Layout name to use when none is specified |
| `editor` | Editor for `layouts config` (falls back to `$EDITOR`, then `nvim`) |

## Commands

```sh
layouts apply              # pick layout via fzf, apply to current session
layouts apply dev          # apply named layout
layouts apply dev -d .     # apply using specific working directory
layouts split              # show split syntax and examples
layouts split 22           # split current pane into 2 rows, 2 columns
layouts split 23           # split current pane into 2 rows, 3 columns
layouts split c22          # relayout the current window into 2 rows, 2 columns
layouts split c23          # relayout the current window into 2 rows, 3 columns
layouts split c34          # relayout the current window into 3 rows, 4 columns
layouts maximize           # maximize current pane, run again to restore
layouts z                  # same toggle, short alias

layouts list               # list all layouts with window/pane counts
layouts show dev           # show layout tree with panes, sizes, commands

layouts new mysession dev  # create new tmux session with layout
layouts new mysession      # create session with default layout (if set)

layouts config             # open config in editor
layouts config --path      # print config file path
layouts init               # create config with example layouts

layouts --version          # print version
```

Most commands have short aliases: `apply`→`a`, `split`→`sp`, `maximize`→`max`/`m`/`zoom`/`z`, `list`→`ls`/`l`, `show`→`s`, `new`→`n`, `config`→`c`/`cfg`.

## Show Output

`layouts show dev` renders a tree view:

```
dev

  window 1: claude
    split: horizontal
    ├ claude [25%] → claude --dangerously-skip-permissions
    ├ editor [50%] → nvim .
    └ shell [25%]

  window 2: test
    split: horizontal
    ├ test-1 [50%]
    └ test-2 [50%]
```

## Fish Alias

The `ly` function maps subcommands to `layouts`:

```sh
ly              # layouts list
ly a dev        # layouts apply dev
ly sp           # show split syntax and examples
ly sp 22        # split current pane into 2 rows, 2 columns
ly sp c23       # relayout the current window into 2 rows, 3 columns
ly z            # maximize current pane, run again to restore
ly s dev        # layouts show dev
ly n work dev   # layouts new work dev
ly c            # layouts config
```

## Split Specs

`layouts split` can work at two levels. Plain specs act on the **current pane**. Specs with a `c` prefix relayout the **current tmux window**. This lets you either expand just the active pane or normalize the whole window into a clean grid.

Specs are generated at runtime:

- `<rows><cols>` creates a grid inside the current pane. Example: `22`, `34`
- `<rows>x<cols>` does the same thing with clearer multi-digit syntax. Example: `3x4`
- `c<spec>` relayouts the **current tmux window** into that grid and keeps the existing panes inside the new cells. Example: `c23`, `c3x4`
- Generated grid panes are titled as `<window>.<row>.<column> <dog-alias>`. For example, row 1 column 2 in tmux window `2` becomes `2.1.2 beagle`.

Examples:

- `12` or `2cols` — 1 row, 2 columns
- `23` — 2 rows, 3 columns
- `22` or `a22` — 2 rows, 2 columns
- `c12` — relayout the current window into 1 row, 2 columns
- `c22` or `c` or `current` — relayout the current window into 2 rows, 2 columns
- `c23` — relayout the current window into 2 rows, 3 columns

## How It Works

`layouts apply` adds new windows to your **current** tmux session. It does not touch existing windows — it only creates new ones. Each window is split according to the layout spec, and pane commands are sent via `tmux send-keys`.

`layouts new` creates a **new** tmux session with the layout pre-applied. The first window reuses the session's initial window (renamed), subsequent windows are created fresh.

`layouts split` keeps the current pane selected after creating the new panes so the workflow stays predictable. Plain specs like `23` only split the **currently focused pane**. Specs with a `c` prefix relayout the **whole current window** into the requested grid and move any existing panes into the new cells before filling the remaining slots.

After a split, `layouts split` enables pane border titles for that tmux window and names every generated grid cell with its window, row, column, and a short dog-breed alias. The title is stored in a layout-manager pane option, so shell or program title updates do not overwrite it. The aliases are hardcoded and deterministic, so the same cell position gets the same simple name each time.

`layouts maximize` uses tmux's built-in zoom toggle for the **currently focused pane**. Running it again restores the original split view.

Pane sizes are computed proportionally. If some panes have explicit sizes and others don't, the remaining space is divided equally among unspecified panes. Sizes must sum to at most 100%.

## Integration with Grove

If you use [grove](https://github.com/...) for worktree management, you can reference layouts by name in grove's repo config:

```yaml
# ~/.config/grove/config.yaml
repos:
  - path: ~/code/myproject
    layout: dev    # references a layout defined in layouts config
```

---

> Personal tool built for my own workflow. Feel free to fork and adapt.
