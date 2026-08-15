#!/usr/bin/env python3
"""
Navitone Omarchy Quattro Theme Color Extractor
Extracts colors from current Omarchy/Omarchy Quattro theme and outputs TOML config for Navitone.
Supports Omarchy 4.0 Quattro 24-color palettes, Base24 schemes, and Cava theme gradient colors.
"""

import json
import os
import sys
import re
from pathlib import Path

def simple_toml_dump(data, f):
    """Simple TOML writer for our specific use case."""
    for section_name, section_data in data.items():
        f.write(f"[{section_name}]\n")
        if isinstance(section_data, dict):
            for key, value in section_data.items():
                if isinstance(value, dict):
                    f.write(f"\n[{section_name}.{key}]\n")
                    for nested_key, nested_value in value.items():
                        if isinstance(nested_value, list):
                            formatted_list = ", ".join(f'"{v}"' for v in nested_value)
                            f.write(f'{nested_key} = [{formatted_list}]\n')
                        elif isinstance(nested_value, str):
                            f.write(f'{nested_key} = "{nested_value}"\n')
                        else:
                            f.write(f'{nested_key} = {nested_value}\n')
                elif isinstance(value, list):
                    formatted_list = ", ".join(f'"{v}"' for v in value)
                    f.write(f'{key} = [{formatted_list}]\n')
                elif isinstance(value, str):
                    f.write(f'{key} = "{value}"\n')
                else:
                    f.write(f'{key} = {value}\n')
        f.write("\n")

def get_current_theme_name():
    """Get the name of the current Omarchy theme."""
    home = Path.home()
    theme_name_files = [
        home / '.local' / 'state' / 'omarchy' / 'current' / 'theme.name',
        home / '.local' / 'state' / 'omarchy' / 'theme.name',
        home / '.local' / 'state' / 'omarchy' / 'active_theme',
        home / '.local' / 'state' / 'omarchy' / 'current' / 'theme' / 'theme.name',
        home / '.config' / 'omarchy' / 'current' / 'theme.name',
        home / '.config' / 'omarchy' / 'theme.name',
        home / '.config' / 'omarchy' / 'active_theme',
        home / '.config' / 'omarchy' / 'current' / 'theme' / 'theme.name',
    ]

    for f in theme_name_files:
        if f.exists():
            try:
                name = f.read_text().strip()
                if name:
                    return name
            except Exception:
                pass

    current_theme_path = home / '.local' / 'state' / 'omarchy' / 'current' / 'theme'
    if current_theme_path.exists():
        if current_theme_path.is_symlink():
            return Path(current_theme_path.readlink()).name
        if current_theme_path.is_dir():
            return "current"

    return None

def get_current_theme_path():
    """Get the path to the current Omarchy theme directory."""
    home = Path.home()

    candidates = [
        home / '.local' / 'state' / 'omarchy' / 'current' / 'theme',
        home / '.local' / 'state' / 'omarchy' / 'current',
        home / '.config' / 'omarchy' / 'current' / 'theme',
        home / '.config' / 'omarchy' / 'current',
        home / '.config' / 'omarchy' / 'theme',
    ]

    for cand in candidates:
        if cand.exists():
            if cand.is_symlink():
                return Path(cand.readlink())
            if cand.is_dir():
                return cand

    theme_name = get_current_theme_name()
    if theme_name:
        named_paths = [
            home / '.config' / 'omarchy' / 'themes' / theme_name,
            home / '.local' / 'share' / 'omarchy' / 'themes' / theme_name,
        ]
        for p in named_paths:
            if p.exists() and p.is_dir():
                return p

    themes_dir = home / '.config' / 'omarchy' / 'themes'
    if themes_dir.exists() and themes_dir.is_dir():
        for t_dir in themes_dir.iterdir():
            if t_dir.is_dir() and ((t_dir / 'colors.toml').exists() or (t_dir / 'custom_theme.json').exists() or (t_dir / 'alacritty.toml').exists()):
                return t_dir

    return None

def parse_colors_toml(colors_path):
    """Parse Omarchy Quattro colors.toml format including 24 colors & Base24 schemes."""
    raw_colors = {}
    background = '#282a36'
    foreground = '#f8f8f2'
    color_indices = {}
    base_indices = {}

    try:
        with open(colors_path, 'r', encoding='utf-8') as f:
            content = f.read()

        for line in content.split('\n'):
            line = line.strip()
            if not line or line.startswith('#') or (line.startswith('[') and line.endswith(']')):
                continue

            if '=' in line:
                key, value = line.split('=', 1)
                key = key.strip()
                if '.' in key:
                    key = key.split('.')[-1].strip()

                hex_match = re.search(r'(?:#|0x)([a-fA-F0-9]{6}|[a-fA-F0-9]{3})\b', value)
                if hex_match:
                    hex_str = hex_match.group(1).upper()
                    if len(hex_str) == 3:
                        clean_color = '#' + ''.join(c * 2 for c in hex_str)
                    else:
                        clean_color = '#' + hex_str

                    raw_colors[key] = clean_color

                    if key == 'background':
                        background = clean_color
                    elif key == 'foreground':
                        foreground = clean_color

                    if key.startswith('color') and key[5:].isdigit():
                        idx = int(key[5:])
                        color_indices[idx] = clean_color

                    if key.startswith('base') and len(key) == 6:
                        try:
                            idx = int(key[4:], 16)
                            base_indices[idx] = clean_color
                        except ValueError:
                            pass

        palette = []
        if color_indices:
            for k in sorted(color_indices.keys()):
                palette.append(color_indices[k])
        elif base_indices:
            for k in sorted(base_indices.keys()):
                palette.append(base_indices[k])

        if len(palette) <= 1:
            for k in ['accent', 'primary', 'secondary', 'success', 'warning', 'error']:
                if k in raw_colors and raw_colors[k] not in palette:
                    palette.append(raw_colors[k])

        accent = raw_colors.get('accent', raw_colors.get('color4', raw_colors.get('color12', '#6272a4')))
        primary = raw_colors.get('primary', raw_colors.get('color4', raw_colors.get('color6', accent)))
        secondary = raw_colors.get('secondary', raw_colors.get('color5', raw_colors.get('color13', '#ff79c6')))
        success = raw_colors.get('success', raw_colors.get('color2', raw_colors.get('color10', '#50fa7b')))
        warning = raw_colors.get('warning', raw_colors.get('color3', raw_colors.get('color11', '#f1fa8c')))
        error = raw_colors.get('error', raw_colors.get('color1', raw_colors.get('color9', '#ff5555')))

        colors = {
            'accent': accent,
            'primary': primary,
            'secondary': secondary,
            'success': success,
            'warning': warning,
            'error': error,
        }

        return colors, palette, background, foreground

    except Exception as e:
        print(f"Error parsing colors.toml: {e}")
        return None, [], background, foreground

def parse_cava_theme(theme_path):
    """Parse gradient_color_1 .. gradient_color_N from cava_theme."""
    candidates = [
        theme_path / 'cava_theme',
        theme_path / 'theme' / 'cava_theme',
        Path.home() / '.local' / 'state' / 'omarchy' / 'current' / 'theme' / 'cava_theme',
        Path.home() / '.config' / 'omarchy' / 'current' / 'theme' / 'cava_theme',
    ]
    for c_path in candidates:
        if c_path.exists():
            try:
                with open(c_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                grad_map = {}
                for line in content.split('\n'):
                    line = line.strip()
                    if line.startswith('#') or line.startswith(';'):
                        continue
                    if '=' in line and ('#' in line or '0x' in line):
                        k, v = line.split('=', 1)
                        k = k.strip()
                        if 'gradient_color_' in k:
                            num_str = k.split('gradient_color_')[-1].strip()
                            if num_str.isdigit():
                                m = re.search(r'(?:#|0x)([a-fA-F0-9]{6}|[a-fA-F0-9]{3})\b', v)
                                if m:
                                    hex_str = m.group(1).upper()
                                    clean_color = '#' + (''.join(c * 2 for c in hex_str) if len(hex_str) == 3 else hex_str)
                                    grad_map[int(num_str)] = clean_color
                if grad_map:
                    return [grad_map[k] for k in sorted(grad_map.keys())]
            except Exception:
                pass
    return []

def extract_theme_colors(theme_path):
    """Extract relevant colors from the Omarchy theme."""
    theme_name = get_current_theme_name() or theme_path.name
    cava_palette = parse_cava_theme(theme_path)

    # 1. Try colors.toml (Omarchy/Omarchy Quattro format)
    colors_toml_candidates = [
        theme_path / 'colors.toml',
        theme_path / 'theme' / 'colors.toml',
    ]

    for colors_toml_path in colors_toml_candidates:
        if colors_toml_path.exists():
            colors, palette, background, foreground = parse_colors_toml(colors_toml_path)
            if colors:
                vibrant_cava = [c for c in palette if c.upper() != background.upper()] if palette else []
                return {
                    'theme_name': theme_name,
                    'colors': colors,
                    'palette': palette,
                    'cava_palette': cava_palette or (vibrant_cava if vibrant_cava else palette),
                    'background': background,
                    'foreground': foreground
                }

    # 2. Try custom_theme.json
    theme_json_path = theme_path / 'custom_theme.json'
    if theme_json_path.exists():
        try:
            with open(theme_json_path, 'r', encoding='utf-8') as f:
                theme_data = json.load(f)

            colors = {}
            palette = []
            if 'colors' in theme_data and 'terminal' in theme_data['colors']:
                term = theme_data['colors']['terminal']
                colors = {
                    'accent': term.get('blue', '#38d6fa'),
                    'primary': term.get('cyan', '#048ba8'),
                    'secondary': term.get('magenta', '#d35f5f'),
                    'success': term.get('green', '#a9fbd7'),
                    'warning': term.get('yellow', '#9f87af'),
                    'error': term.get('red', '#9c528b'),
                }
                palette = [v for v in term.values() if isinstance(v, str) and v.startswith('#')]

            background = '#282a36'
            foreground = '#f8f8f2'
            if 'colors' in theme_data and 'primary' in theme_data['colors']:
                background = theme_data['colors']['primary'].get('background', background)
                foreground = theme_data['colors']['primary'].get('foreground', foreground)

            if colors:
                return {
                    'theme_name': theme_name,
                    'colors': colors,
                    'palette': palette,
                    'cava_palette': cava_palette or palette,
                    'background': background,
                    'foreground': foreground
                }
        except Exception as e:
            print(f"Error reading JSON theme: {e}")

    # 3. Fallback to alacritty.toml
    alacritty_toml = theme_path / 'alacritty.toml'
    if alacritty_toml.exists():
        try:
            with open(alacritty_toml, 'r', encoding='utf-8') as f:
                content = f.read()

            colors_map = {}
            for line in content.split('\n'):
                line = line.strip()
                if '=' in line and ('#' in line or '0x' in line):
                    k, v = line.split('=', 1)
                    k = k.strip().split('.')[-1]
                    m = re.search(r'(?:#|0x)([a-fA-F0-9]{6})', v)
                    if m:
                        colors_map[k] = '#' + m.group(1)

            colors = {
                'accent': colors_map.get('blue', '#6272a4'),
                'primary': colors_map.get('cyan', '#8be9fd'),
                'secondary': colors_map.get('magenta', '#ff79c6'),
                'success': colors_map.get('green', '#50fa7b'),
                'warning': colors_map.get('yellow', '#f1fa8c'),
                'error': colors_map.get('red', '#ff5555'),
            }
            background = colors_map.get('background', '#282a36')
            foreground = colors_map.get('foreground', '#f8f8f2')

            return {
                'theme_name': theme_name,
                'colors': colors,
                'palette': list(colors.values()),
                'cava_palette': cava_palette or list(colors.values()),
                'background': background,
                'foreground': foreground
            }
        except Exception as e:
            print(f"Error reading alacritty.toml theme: {e}")

    return None

def generate_navitone_config(theme_data):
    """Generate Navitone theme configuration."""
    colors_dict = theme_data['colors'].copy()
    if theme_data.get('palette'):
        colors_dict['palette'] = theme_data['palette']
    if theme_data.get('cava_palette'):
        colors_dict['cava_palette'] = theme_data['cava_palette']

    config = {
        'theme': {
            'name': f"omarchy-{theme_data['theme_name']}",
            'source': 'omarchy',
            'colors': colors_dict,
            'background': theme_data['background'],
            'foreground': theme_data['foreground']
        }
    }
    return config

def update_navitone_config(theme_config):
    """Update Navitone's config.toml with theme colors - SAFELY preserving existing config."""
    config_path = Path.home() / '.config' / 'navitone-cli' / 'config.toml'

    if not config_path.exists():
        print(f"Navitone config not found at {config_path}")
        return False

    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        new_lines = []
        skip_theme_section = False

        for line in lines:
            line_stripped = line.strip()

            if line_stripped == '[theme]' or line_stripped.startswith('[theme.'):
                skip_theme_section = True
                continue

            if skip_theme_section and line_stripped.startswith('[') and not line_stripped.startswith('[theme'):
                skip_theme_section = False

            if not skip_theme_section:
                new_lines.append(line)

        if new_lines and not new_lines[-1].endswith('\n'):
            new_lines.append('\n')

        new_lines.append('\n[theme]\n')
        new_lines.append(f'name = "{theme_config["theme"]["name"]}"\n')
        new_lines.append(f'source = "{theme_config["theme"]["source"]}"\n')
        new_lines.append('\n[theme.colors]\n')

        for color_name, color_value in theme_config["theme"]["colors"].items():
            if isinstance(color_value, list):
                formatted_list = ", ".join(f'"{v}"' for v in color_value)
                new_lines.append(f'{color_name} = [{formatted_list}]\n')
            else:
                new_lines.append(f'{color_name} = "{color_value}"\n')

        new_lines.append(f'background = "{theme_config["theme"]["background"]}"\n')
        new_lines.append(f'foreground = "{theme_config["theme"]["foreground"]}"\n')
        new_lines.append('\n')

        with open(config_path, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)

        return True

    except Exception as e:
        print(f"Error updating Navitone config: {e}")
        return False

def main():
    """Main function to extract Omarchy colors and update Navitone config."""
    print("Extracting Omarchy Quattro theme colors for Navitone...")

    theme_path = get_current_theme_path()
    if not theme_path:
        print("Could not find current Omarchy theme path")
        sys.exit(1)

    theme_name = get_current_theme_name() or theme_path.name
    print(f"Current Omarchy theme: {theme_name}")

    theme_data = extract_theme_colors(theme_path)
    if not theme_data:
        print("Could not extract colors from theme")
        sys.exit(1)

    print(f"Extracted colors: {theme_data['colors']}")
    if theme_data.get('palette'):
        print(f"Extracted Omarchy Quattro palette ({len(theme_data['palette'])} colors)")
    if theme_data.get('cava_palette'):
        print(f"Extracted Omarchy Cava palette ({len(theme_data['cava_palette'])} gradient colors)")

    navitone_config = generate_navitone_config(theme_data)

    if update_navitone_config(navitone_config):
        print("✅ Navitone config updated successfully!")
        print(f"Theme '{navitone_config['theme']['name']}' applied")
    else:
        print("❌ Failed to update Navitone config")
        sys.exit(1)

if __name__ == '__main__':
    main()