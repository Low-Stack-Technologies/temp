#!/bin/bash

# Function to detect OS and architecture
get_os_arch() {
    local OS
    local ARCH

    # Detect OS
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        MINGW*)     OS="windows";;
        *)          echo "Unsupported OS" && exit 1;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64*)    ARCH="amd64";;
        arm64*)     ARCH="arm64";;
        aarch64*)   ARCH="arm64";;
        *)          echo "Unsupported architecture" && exit 1;;
    esac

    echo "${OS}-${ARCH}"
}

# Get OS and architecture
OS_ARCH=$(get_os_arch)
BASE_URL="https://github.com/low-stack/temp/releases/latest/download/temp"

case $OS_ARCH in
    "linux-amd64"|"linux-arm64")
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"

        # Download binary
        curl -L "${BASE_URL}-${OS_ARCH}" -o "$INSTALL_DIR/temp"
        chmod +x "$INSTALL_DIR/temp"

        # Add to PATH if not already present
        for RC_FILE in ".bashrc" ".zshrc"; do
            if [ -f "$HOME/$RC_FILE" ]; then
                if ! grep -q "$INSTALL_DIR" "$HOME/$RC_FILE"; then
                    echo "export PATH=\$PATH:$INSTALL_DIR" >> "$HOME/$RC_FILE"
                    echo "Updated $RC_FILE"
                fi
            fi
        done
        echo "Please restart your shell or source your RC file to update PATH"
        ;;

    "darwin-amd64"|"darwin-arm64")
        INSTALL_DIR="/usr/local/bin"

        # Download binary
        sudo curl -L "${BASE_URL}-${OS_ARCH}" -o "$INSTALL_DIR/temp"
        sudo chmod +x "$INSTALL_DIR/temp"
        echo "Installation complete - binary installed to $INSTALL_DIR"
        ;;

    "windows-amd64"|"windows-arm64")
        INSTALL_DIR="$HOME/AppData/Local/temp"
        mkdir -p "$INSTALL_DIR"

        # Download binary
        curl -L "${BASE_URL}-${OS_ARCH}.exe" -o "$INSTALL_DIR/temp.exe"

        # Add to PATH using PowerShell
        powershell.exe -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$INSTALL_DIR', 'User')"
        echo "Installation complete - please restart your terminal"
        ;;
esac