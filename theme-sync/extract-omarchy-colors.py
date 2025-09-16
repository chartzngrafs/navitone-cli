#!/usr/bin/env python3
"""
Navitone Omarchy Theme Color Extractor
Extracts colors from current Omarchy theme and outputs TOML config for Navitone.
"""

import json
import os
import sys
import re
from pathlib import Path

# Simple TOML writer since toml module might not be available
def simple_toml_dump(data, f):
    """Simple TOML writer for our specific use case."""
    for section_name, section_data in data.items():
        f.write(f"[{section_name}]\n")
        if isinstance(section_data, dict):
            for key, value in section_data.items():
                if isinstance(value, dict):
                    # Nested section
                    f.write(f"\n[{section_name}.{key}]\n")
                    for nested_key, nested_value in value.items():
                        if isinstance(nested_value, str):
                            f.write(f'{nested_key} = "{nested_value}"\n')
                        else:
                            f.write(f'{nested_key} = {nested_value}\n')
                elif isinstance(value, str):
                    f.write(f'{key} = "{value}"\n')
                else:
                    f.write(f'{key} = {value}\n')
        f.write("\n")

def simple_toml_load(f):
    """Simple TOML loader for our specific use case."""
    content = f.read()
    data = {}
    current_section = None
    current_nested = None

    for line in content.split('\n'):
        line = line.strip()
        if not line or line.startswith('#'):
            continue

        # Section headers
        if line.startswith('[') and line.endswith(']'):
            section_name = line[1:-1]
            if '.' in section_name:
                parts = section_name.split('.', 1)
                current_section = parts[0]
                current_nested = parts[1]
                if current_section not in data:
                    data[current_section] = {}
                data[current_section][current_nested] = {}
            else:
                current_section = section_name
                current_nested = None
                data[current_section] = {}
            continue

        # Key-value pairs
        if '=' in line and current_section:
            key, value = line.split('=', 1)
            key = key.strip()
            value = value.strip().strip('"\'')

            if current_nested:
                data[current_section][current_nested][key] = value
            else:
                data[current_section][key] = value

    return data

def get_current_theme_path():
    """Get the path to the current Omarchy theme."""
    current_theme_link = Path.home() / '.config' / 'omarchy' / 'current' / 'theme'
    if current_theme_link.exists() and current_theme_link.is_symlink():
        return Path(current_theme_link.readlink())
    return None

def parse_toml_colors(toml_path):
    """Parse colors from TOML file with comprehensive parsing like Cava app."""
    normal_colors = {}
    bright_colors = {}
    primary_colors = {}
    current_section = None

    try:
        with open(toml_path, 'r') as f:
            content = f.read()

        # Parse line by line, tracking sections
        for line in content.split('\n'):
            line = line.strip()

            # Track TOML sections
            if line.startswith('[colors.normal]'):
                current_section = 'normal'
                continue
            elif line.startswith('[colors.bright]'):
                current_section = 'bright'
                continue
            elif line.startswith('[colors.primary]'):
                current_section = 'primary'
                continue
            elif line.startswith('[') and 'colors' not in line:
                current_section = None  # Other section
                continue
            elif line.startswith('[colors.'):
                current_section = 'other'  # Other color section
                continue

            # Skip commented lines
            if line.startswith('#'):
                continue

            # Look for color assignments (handle inline comments)
            if '=' in line and ('0x' in line or '#' in line):
                key, value = line.split('=', 1)
                key = key.strip()
                value = value.strip().strip('"\'')

                # Extract hex color in either #xxxxxx or 0xxxxxxx format
                # Handle inline comments by taking first color found
                hex_match = re.search(r'(?:#|0x)([a-fA-F0-9]{6})', value)
                if hex_match:
                    # Convert to standard #xxxxxx format
                    clean_color = '#' + hex_match.group(1)

                    # Extract just the color name
                    color_name = key.split('.')[-1] if '.' in key else key

                    # Store in appropriate section
                    if current_section == 'normal':
                        normal_colors[color_name] = clean_color
                    elif current_section == 'bright':
                        bright_colors[color_name] = clean_color
                    elif current_section == 'primary':
                        primary_colors[color_name] = clean_color

    except Exception as e:
        print(f"Error parsing TOML: {e}")

    # Prefer normal colors over bright colors for a more muted palette (like Cava)
    colors = normal_colors.copy()

    # Fill in missing colors with bright variants if needed
    for color_name in ['red', 'green', 'yellow', 'blue', 'magenta', 'cyan']:
        if color_name not in colors and color_name in bright_colors:
            colors[color_name] = bright_colors[color_name]

    # Debug output when both types are found
    if normal_colors and bright_colors:
        print(f"Found both normal and bright colors, prioritizing normal colors")

    return colors, primary_colors

def extract_theme_colors(theme_path):
    """Extract relevant colors from the Omarchy theme."""
    theme_name = theme_path.name

    # First try to read from custom_theme.json
    theme_json_path = theme_path / 'custom_theme.json'
    if theme_json_path.exists():
        try:
            with open(theme_json_path, 'r') as f:
                theme_data = json.load(f)

            colors = {}

            # Try to get terminal colors first
            if 'colors' in theme_data and 'terminal' in theme_data['colors']:
                terminal_colors = theme_data['colors']['terminal']
                colors = {
                    'accent': terminal_colors.get('blue', '#38d6fa'),
                    'primary': terminal_colors.get('cyan', '#048ba8'),
                    'secondary': terminal_colors.get('magenta', '#d35f5f'),
                    'success': terminal_colors.get('green', '#a9fbd7'),
                    'warning': terminal_colors.get('yellow', '#9f87af'),
                    'error': terminal_colors.get('red', '#9c528b'),
                }
            # Fallback to alacritty normal colors
            elif 'apps' in theme_data and 'alacritty' in theme_data['apps']:
                alacritty_colors = theme_data['apps']['alacritty']['colors']['normal']
                colors = {
                    'accent': alacritty_colors.get('blue', '#38d6fa'),
                    'primary': alacritty_colors.get('cyan', '#048ba8'),
                    'secondary': alacritty_colors.get('magenta', '#d35f5f'),
                    'success': alacritty_colors.get('green', '#a9fbd7'),
                    'warning': alacritty_colors.get('yellow', '#9f87af'),
                    'error': alacritty_colors.get('red', '#9c528b'),
                }

            # Get background and foreground
            background = '#000000'
            foreground = '#ffffff'
            if 'colors' in theme_data and 'primary' in theme_data['colors']:
                background = theme_data['colors']['primary'].get('background', background)
                foreground = theme_data['colors']['primary'].get('foreground', foreground)
            elif 'apps' in theme_data and 'alacritty' in theme_data['apps']:
                primary = theme_data['apps']['alacritty']['colors']['primary']
                background = primary.get('background', background)
                foreground = primary.get('foreground', foreground)

            return {
                'theme_name': theme_name,
                'colors': colors,
                'background': background,
                'foreground': foreground
            }

        except (json.JSONDecodeError, KeyError) as e:
            print(f"Error reading JSON theme: {e}")

    # Fallback: try to read from alacritty.toml
    alacritty_toml = theme_path / 'alacritty.toml'
    if alacritty_toml.exists():
        try:
            toml_colors, primary_colors = parse_toml_colors(alacritty_toml)

            # Check if we actually found any colors, if not, create a monochrome scheme (like Cava)
            if not toml_colors:
                # Try to extract foreground color from the file
                foreground = '#EFEFEF'  # default
                try:
                    with open(alacritty_toml, 'r') as f:
                        content = f.read()
                    # Look for foreground in various formats
                    fg_match = re.search(r'foreground\s*=\s*["\']?(?:#|0x)([a-fA-F0-9]{6})["\']?', content)
                    if fg_match:
                        foreground = '#' + fg_match.group(1)
                except:
                    pass

                # Create monochrome gradient from dark to light (like midnight theme)
                print(f"Theme has no color palette, using monochrome scheme with foreground: {foreground}")
                colors = {
                    'accent': '#777777',      # Medium gray for headers
                    'primary': '#999999',     # Light gray for primary elements
                    'secondary': '#BBBBBB',   # Lighter gray for selections
                    'success': '#AAAAAA',     # Success in grayscale
                    'warning': '#888888',     # Warning in grayscale
                    'error': '#666666',       # Error in darker gray
                }
            else:
                # Map available colors to Navitone color scheme
                colors = {
                    'accent': toml_colors.get('blue', '#6272a4'),
                    'primary': toml_colors.get('cyan', '#8be9fd'),
                    'secondary': toml_colors.get('magenta', '#ff79c6'),
                    'success': toml_colors.get('green', '#50fa7b'),
                    'warning': toml_colors.get('yellow', '#f1fa8c'),
                    'error': toml_colors.get('red', '#ff5555'),
                }

            # Get background and foreground from primary colors or defaults
            background = primary_colors.get('background', '#282a36')
            foreground = primary_colors.get('foreground', '#f8f8f2')

            # If no primary colors found, try to extract from content
            if not primary_colors:
                try:
                    with open(alacritty_toml, 'r') as f:
                        content = f.read()

                    # Extract background
                    bg_match = re.search(r'background\s*=\s*["\']?(?:#|0x)([a-fA-F0-9]{6})["\']?', content)
                    if bg_match:
                        background = '#' + bg_match.group(1)

                    # Extract foreground
                    fg_match = re.search(r'foreground\s*=\s*["\']?(?:#|0x)([a-fA-F0-9]{6})["\']?', content)
                    if fg_match:
                        foreground = '#' + fg_match.group(1)
                except:
                    pass

            return {
                'theme_name': theme_name,
                'colors': colors,
                'background': background,
                'foreground': foreground
            }

        except Exception as e:
            print(f"Error reading TOML theme: {e}")

    return None

def generate_navitone_config(theme_data):
    """Generate Navitone theme configuration."""
    config = {
        'theme': {
            'name': f"omarchy-{theme_data['theme_name']}",
            'source': 'omarchy',
            'colors': theme_data['colors'],
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
        # Read the current config file as text to preserve formatting and comments
        with open(config_path, 'r') as f:
            lines = f.readlines()

        # Find existing [theme] section and remove it
        new_lines = []
        skip_theme_section = False

        for line in lines:
            line_stripped = line.strip()

            # Check if we're entering the theme section or theme subsections
            if line_stripped == '[theme]' or line_stripped.startswith('[theme.'):
                skip_theme_section = True
                continue

            # Check if we're entering a new section (exit theme section)
            if skip_theme_section and line_stripped.startswith('[') and not line_stripped.startswith('[theme'):
                skip_theme_section = False

            # Skip lines in theme section, but include lines from other sections
            if not skip_theme_section:
                new_lines.append(line)

        # Add the new theme section at the end
        if new_lines and not new_lines[-1].endswith('\n'):
            new_lines.append('\n')

        new_lines.append('\n[theme]\n')
        new_lines.append(f'name = "{theme_config["theme"]["name"]}"\n')
        new_lines.append(f'source = "{theme_config["theme"]["source"]}"\n')
        new_lines.append('\n[theme.colors]\n')

        for color_name, color_value in theme_config["theme"]["colors"].items():
            new_lines.append(f'{color_name} = "{color_value}"\n')

        new_lines.append(f'background = "{theme_config["theme"]["background"]}"\n')
        new_lines.append(f'foreground = "{theme_config["theme"]["foreground"]}"\n')
        new_lines.append('\n')

        # Write the updated config
        with open(config_path, 'w') as f:
            f.writelines(new_lines)

        return True

    except Exception as e:
        print(f"Error updating Navitone config: {e}")
        return False

def main():
    """Main function to extract Omarchy colors and update Navitone config."""
    print("Extracting Omarchy theme colors for Navitone...")

    # Get current theme path
    theme_path = get_current_theme_path()
    if not theme_path:
        print("Could not find current Omarchy theme")
        sys.exit(1)

    theme_name = theme_path.name
    print(f"Current theme: {theme_name}")

    # Extract colors from theme
    theme_data = extract_theme_colors(theme_path)
    if not theme_data:
        print("Could not extract colors from theme")
        sys.exit(1)

    print(f"Extracted colors: {theme_data['colors']}")

    # Generate Navitone config
    navitone_config = generate_navitone_config(theme_data)

    # Output for manual use
    print("\n--- Navitone Theme Config (TOML) ---")
    import io
    output = io.StringIO()
    simple_toml_dump(navitone_config, output)
    print(output.getvalue())

    # Update Navitone config directly
    if update_navitone_config(navitone_config):
        print("✅ Navitone config updated successfully!")
        print(f"Theme '{navitone_config['theme']['name']}' applied")
    else:
        print("❌ Failed to update Navitone config")
        sys.exit(1)

if __name__ == '__main__':
    main()