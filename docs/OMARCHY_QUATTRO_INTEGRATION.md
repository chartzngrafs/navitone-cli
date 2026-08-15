# Omarchy 4.0 Quattro & Vertical Cava Integration Guide

## Overview

Navitone-CLI includes native integration with **Omarchy 4.0 Quattro**, supporting extended 24-color palettes, Base24 schemas, state path resolution (`~/.local/state/omarchy/`), and bottom-to-top vertical Cava visualizer spectrum color interpolation.

---

## 1. Path Resolution Architecture

Omarchy 4.0 Quattro stores runtime state in `~/.local/state/omarchy/`. Navitone prioritizes state paths in the following sequence:

### Theme Colors (`colors.toml`)
1. `~/.local/state/omarchy/current/theme/colors.toml`
2. `~/.local/state/omarchy/current/colors.toml`
3. `~/.local/state/omarchy/theme/colors.toml`
4. `~/.config/omarchy/current/theme/colors.toml`
5. `~/.config/omarchy/themes/<theme>/colors.toml`

### Active Theme Name (`theme.name`)
1. `~/.local/state/omarchy/current/theme.name`
2. `~/.local/state/omarchy/theme.name`
3. `~/.config/omarchy/current/theme.name`

### Cava Theme (`cava_theme`)
1. `~/.local/state/omarchy/current/theme/cava_theme`
2. `~/.local/state/omarchy/current/cava_theme`
3. `~/.config/omarchy/themes/<theme>/cava_theme`

---

## 2. 24-Color & Base24 Palette Loader

`ParseOmarchyColorsTOML()` parses:
- 24-color Quattro keys: `color0` .. `color23`
- Base24 hex keys: `base00` .. `base17`
- Semantic UI roles: `accent`, `primary`, `secondary`, `success`, `warning`, `error`, `background`, `foreground`

If indexed colors (`color0`..`color23`) are absent or single-color, Navitone automatically appends available semantic colors to guarantee a rich multi-stop gradient.

---

## 3. Integrated Vertical Cava Visualizer Spectrum

Cava spectrum bars (`Shift+C`) render using vertical bottom-to-top color gradient interpolation (`GetCavaGradientColor`):
- Row `0` (bottom of bar): Base color (`CavaPalette[0]`)
- Row `N` (top peak of bar): Top color (`CavaPalette[len(palette)-1]`)

### Fallback & Background Filtering
When `cava_theme` is absent, Navitone derives `cava_palette` directly from the active theme's `palette`, filtering out the background color (`bg`) so spectrum bars remain vibrant and clear against dark terminal backgrounds.

---

## 4. Live Hotkey Reload (`Shift+R`)

Pressing `Shift+R` or `R` triggers `reloadOmarchyTheme()`, which re-reads active Omarchy Quattro state files and updates both UI components and Cava visualizer colors live without restarting the player.
