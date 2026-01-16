#!/bin/bash

# setup-deps.sh - Install dependencies for goany graphics runtime
# Supports: macOS, Linux (Debian/Ubuntu, Fedora, Arch), Windows (via MSYS2)

set -e

echo "=== goany Graphics Runtime - Dependency Setup ==="
echo ""

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "macos" ;;
        Linux*)     echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

# Detect Linux distribution
detect_linux_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case "$ID" in
            ubuntu|debian|linuxmint|pop) echo "debian" ;;
            fedora|rhel|centos|rocky|alma) echo "fedora" ;;
            arch|manjaro|endeavouros) echo "arch" ;;
            *) echo "unknown" ;;
        esac
    else
        echo "unknown"
    fi
}

# Install SDL2 on macOS
install_macos() {
    echo "Detected: macOS"
    echo ""

    if ! command -v brew &> /dev/null; then
        echo "Error: Homebrew not found. Please install from https://brew.sh"
        exit 1
    fi

    echo "Installing SDL2 via Homebrew..."
    brew install sdl2

    echo ""
    echo "SDL2 installed successfully!"
    echo "Include path: $(brew --prefix sdl2)/include"
    echo "Library path: $(brew --prefix sdl2)/lib"
}

# Install SDL2 on Linux
install_linux() {
    local distro=$(detect_linux_distro)
    echo "Detected: Linux ($distro)"
    echo ""

    case "$distro" in
        debian)
            echo "Installing SDL2 via apt..."
            sudo apt-get update
            sudo apt-get install -y libsdl2-dev
            ;;
        fedora)
            echo "Installing SDL2 via dnf..."
            sudo dnf install -y SDL2-devel
            ;;
        arch)
            echo "Installing SDL2 via pacman..."
            sudo pacman -S --noconfirm sdl2
            ;;
        *)
            echo "Error: Unsupported Linux distribution."
            echo "Please install SDL2 development libraries manually:"
            echo "  - Debian/Ubuntu: sudo apt install libsdl2-dev"
            echo "  - Fedora: sudo dnf install SDL2-devel"
            echo "  - Arch: sudo pacman -S sdl2"
            exit 1
            ;;
    esac

    echo ""
    echo "SDL2 installed successfully!"
}

# Install SDL2 on Windows (MSYS2)
install_windows() {
    echo "Detected: Windows (MSYS2/MinGW)"
    echo ""

    if ! command -v pacman &> /dev/null; then
        echo "Error: MSYS2 not found."
        echo "Please install MSYS2 from https://www.msys2.org"
        echo "Then run this script from MSYS2 MinGW terminal."
        exit 1
    fi

    echo "Installing SDL2 via pacman..."
    pacman -S --noconfirm mingw-w64-x86_64-SDL2

    echo ""
    echo "SDL2 installed successfully!"
}

# Verify installation
verify_sdl2() {
    echo ""
    echo "Verifying SDL2 installation..."

    if command -v sdl2-config &> /dev/null; then
        echo "SDL2 version: $(sdl2-config --version)"
        echo "Compiler flags: $(sdl2-config --cflags)"
        echo "Linker flags: $(sdl2-config --libs)"
    elif command -v pkg-config &> /dev/null && pkg-config --exists sdl2; then
        echo "SDL2 version: $(pkg-config --modversion sdl2)"
        echo "Compiler flags: $(pkg-config --cflags sdl2)"
        echo "Linker flags: $(pkg-config --libs sdl2)"
    else
        echo "Warning: Could not verify SDL2 installation."
        echo "SDL2 may still be installed correctly."
    fi
}

# Print backend-specific instructions
print_backend_instructions() {
    echo ""
    echo "=== Backend Setup Instructions ==="
    echo ""
    echo "C++ compilation:"
    echo "  g++ -std=c++17 output.cpp \$(sdl2-config --cflags --libs)"
    echo ""
    echo "C# (add NuGet package):"
    echo "  dotnet add package SDL2-CS"
    echo ""
    echo "Rust (add to Cargo.toml):"
    echo '  [dependencies]'
    echo '  sdl2 = "0.36"'
    echo ""
}

# Main
OS=$(detect_os)

case "$OS" in
    macos)   install_macos ;;
    linux)   install_linux ;;
    windows) install_windows ;;
    *)
        echo "Error: Unsupported operating system."
        echo "Please install SDL2 manually from https://www.libsdl.org"
        exit 1
        ;;
esac

verify_sdl2
print_backend_instructions

echo "=== Setup Complete ==="
