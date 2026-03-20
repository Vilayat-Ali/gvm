# 🐹 GVM: Go Version Manager

<div align="center">

![Version](https://img.shields.io/badge/version-2.0.0-blue?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)

**A modern, lightning-fast CLI tool to manage multiple Go versions without breaking a sweat.**

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Commands](#-commands) • [Configuration](#-configuration) • [FAQ](#-faq)

</div>

---

## ✨ Features

- 🚀 **Blazing Fast** - Minimal overhead, maximum speed
- 🔒 **Secure Downloads** - SHA256 checksum verification for every download
- 🎯 **Cross-Platform** - Works on Linux, macOS, and Windows
- 💻 **Modern CLI** - Beautiful, colorful output that developers love
- 📦 **Zero Dependencies** - Just download and run, no extra setup needed
- 🔄 **Seamless Switching** - Switch between Go versions instantly

---

## 🚀 Installation

### Using the Install Script (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/Vilayat-Ali/gvm/main/scripts/install.sh | bash
```

### From Source

```bash
git clone https://github.com/Vilayat-Ali/gvm.git
cd gvm
make build
sudo make install
```

### Manual Installation

1. Download the latest release for your platform from the [Releases page](https://github.com/Vilayat-Ali/gvm/releases)
2. Extract the binary
3. Move it to a directory in your PATH

```bash
# Example for Linux AMD64
sudo mv gvm-linux-amd64 /usr/local/bin/gvm
sudo chmod +x /usr/local/bin/gvm
```

> **Note:** After installation, restart your terminal or run `source ~/.bashrc` (or your shell's config file).

---

## 🎯 Quick Start

```bash
# First-time setup
gvm configure

# See available versions
gvm list

# Download a version
gvm download 1.22.0

# Start using it!
gvm use 1.22.0

# Verify it's working
go version
# Output: go version go1.22.0 linux/amd64
```

---

## 📚 Commands

### `gvm configure`

Initialize gvm for first-time use. Creates the necessary config files and directories.

```bash
gvm configure
```

**What it does:**
- Creates config at `~/.config/gvm/config.json`
- Sets up version storage at `/usr/local/gvm/go-versions/`
- Fetches the latest available Go versions

---

### `gvm list`

Browse available and installed Go versions.

```bash
gvm list              # Show available versions
gvm list -d           # Show downloaded/installed versions
gvm list -c           # Show current active version
gvm list update       # Refresh the version list
```

**Output example:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          📋 AVAILABLE GO VERSIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  go1.22.0 ✨ LTS
  go1.21.5 •  AVAILABLE
  go1.21.4 •  AVAILABLE
  go1.21.3 •  AVAILABLE

Quick actions:
  gvm download go1.22.0 → grab the LTS version
```

---

### `gvm download`

Download a Go version to your local storage.

```bash
gvm download 1.22.0
```

**Features:**
- Shows download progress
- Verifies SHA256 checksum after download
- Caches for instant switching later

**Output example:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          📥 DOWNLOADING GO 1.22.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Platform: linux/amd64
  downloading... 100%|████████████| 102MB
  ✓ Checksum verified

✅ Download complete!
  📍 Location: /usr/local/gvm/go-versions/go1.22.0.tar.gz

Next: Run 'gvm use 1.22.0' to start using it!
```

---

### `gvm use`

Activate a Go version. If not downloaded, gvm will grab it for you!

```bash
gvm use 1.22.0
```

**Output example:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
          ⚙️  SWITCHING TO GO 1.22.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Removing old installation...
  Extracting new version...

✅ Switched to Go 1.22.0!

🎉 Ready to code! Run 'go version' to verify
```

> **Note:** Make sure `/usr/local/go/bin` is in your PATH. If not, gvm will remind you to add it!

---

### `gvm delete`

Remove a downloaded Go version.

```bash
gvm delete 1.21.0
```

---

## ⚙️ Configuration

### Config Location
- **Config file:** `~/.config/gvm/config.json`
- **Version storage:** `/usr/local/gvm/go-versions/`

### Manual PATH Setup

Add this to your shell config (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin
```

---

## 🔧 Development

### Building from Source

```bash
# Clone the repo
git clone https://github.com/Vilayat-Ali/gvm.git
cd gvm

# Build
make build

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build for current platform |
| `make build-all` | Build for Linux, macOS, Windows |
| `make install` | Install to system |
| `make test` | Run tests |
| `make fmt` | Format code |
| `make lint` | Lint code |
| `make clean` | Remove build artifacts |

---

## ❓ FAQ

### Q: Why do I need to configure PATH?

GVM installs Go versions to `/usr/local/go`, but your shell needs to know where to find it. Adding `/usr/local/go/bin` to your PATH tells your terminal to use the Go installation managed by gvm.

### Q: Can I have multiple versions installed?

Yes! Download as many versions as you want. Use `gvm list -d` to see all downloaded versions, and `gvm use <version>` to switch between them.

### Q: Does gvm verify downloads?

Absolutely! Every download is verified with SHA256 checksums fetched from the official Go servers at go.dev.

### Q: Can I use gvm without root/sudo?

Currently, gvm requires root access to install Go versions to `/usr/local/`. Future versions may support user-level installations.

### Q: What Go versions are supported?

gvm supports all Go versions from 1.17 onwards, including release candidates (rc) and beta versions.

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/your-username/gvm.git`
3. **Create** a feature branch: `git checkout -b feature/amazing-feature`
4. **Make** your changes and test: `make test`
5. **Commit** your changes: `git commit -m 'Add amazing feature'`
6. **Push** to your fork: `git push origin feature/amazing-feature`
7. Open a **Pull Request** on GitHub

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by [nvm](https://github.com/nvm-sh/nvm) and [fnm](https://github.com/Schniz/fnm)
- Powered by [Cobra](https://github.com/spf13/cobra) CLI framework
- Go downloads served by [go.dev](https://go.dev/)

---

<div align="center">

**Built with 💜 for the Go community**

</div>
