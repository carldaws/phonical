# Phonical - Phonics Learning Tool

A system-wide phonics learning application that plays letter sounds as children type, helping them learn letter-sound associations naturally through any application.

## Features

- **System-wide capture**: Works across all applications - browser, text editor, games, etc.
- **Instant feedback**: Plays phonetic sounds immediately when letters are typed
- **Cross-platform**: Supports macOS, Linux, and Windows
- **Lightweight**: Minimal CPU and memory usage
- **Educational**: Helps children learn phonics through regular computer use
- **British English sounds included**: Ready to use with embedded phonetic sounds

## Requirements

- Go 1.21 or later
- System permissions for accessibility/input monitoring

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/phonical.git
cd phonical
```

2. Install Go if you haven't already:
   - Download from https://go.dev/dl/
   - Follow installation instructions for your OS

3. Build the application:
```bash
go mod download
go build -o phonical
```

## Usage

Run the application:
```bash
./phonical
```

For verbose output (debugging):
```bash
./phonical --verbose
```

### macOS Permissions

On first run, you'll need to grant Accessibility permissions:
1. System will prompt for permission, or
2. Go to System Preferences → Security & Privacy → Privacy → Accessibility
3. Add your Terminal application (Terminal, iTerm2, etc.)
4. Restart the terminal after granting permissions

### Linux Permissions

May require running with elevated permissions or adding your user to the `input` group:
```bash
sudo usermod -a -G input $USER
# Log out and back in for changes to take effect
```

### Running as a Background Service

You can configure Phonical to run automatically at startup using your system's service manager (launchd on macOS, systemd on Linux, Task Scheduler on Windows). The specifics will depend on your operating system and preferences.

## How It Works

Phonical uses system-level keyboard hooks to monitor keystrokes across all applications. When a letter key is pressed, it instantly plays the corresponding phonetic sound file, helping children associate letters with their sounds during normal computer use.

The application:
- Preloads all sound files at startup for instant playback
- Queues sounds to play sequentially if multiple keys are pressed quickly
- Uses minimal system resources
- Respects system audio settings

## Building from Source

```bash
# Clone repository
git clone https://github.com/yourusername/phonical.git
cd phonical

# Download dependencies
go mod download

# Build for current platform
go build -o phonical

# Cross-compile examples
GOOS=darwin GOARCH=amd64 go build -o phonical-mac
GOOS=linux GOARCH=amd64 go build -o phonical-linux
GOOS=windows GOARCH=amd64 go build -o phonical.exe
```

## Sound Files

The repository includes British English phonetic sounds for all letters A-Z in the `sounds/` directory. These are embedded in the binary during compilation, so no external sound files are needed.

### Using Custom Sounds

If you want to use different sounds (e.g., American English pronunciation or different phonetic style):
1. Replace the WAV files in the `sounds/` directory
2. Keep the same naming convention: `a.wav`, `b.wav`, ... `z.wav`
3. Rebuild the application

The application works best with 44.1kHz stereo WAV files.

## Troubleshooting

- **No sound playing**: Check system audio is working and volume is up
- **Permission denied**: Grant accessibility/input permissions as described above
- **High CPU usage**: Run with `--verbose` to check for errors
- **Sounds cutting off**: Ensure WAV files are properly formatted (44.1kHz recommended)

## License

MIT