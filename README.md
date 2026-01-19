# 🐹 GVM: Go Version Manager

**GVM** is a simple, lightweight command-line tool that allows you to manage multiple active Go installations. Switch between versions seamlessly, test your projects against different Go releases, and keep your development environment clean.

## ✨ Features

* **Zero Dependencies:** Built to be fast and portable.
* **Familiar Syntax:** If you've used `nvm` or `fnm`, you already know how to use `gvm`.
* **Version Switching:** Change your global or shell-specific Go version instantly.
* **Auto-Completion:** Full shell completion support for Bash, Zsh, and Fish.

---

## 🚀 Installation

### Using Curl (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/Vilayat-Ali/gvm/main/scripts/install.sh | bash 
```

> **Note:** After installation, remember to restart your terminal or source your config file (e.g., `source ~/.zshrc`).

---

## 🛠 Usage

GVM aims to keep commands intuitive and minimal.

| Command | Description |
| --- | --- |
| `gvm install <version>` | Download and install a specific Go version |
| `gvm use <version>` | Switch the current shell to use `<version>` |
| `gvm list` | List all installed versions |
| `gvm list-remote` | List all available versions from Go official |
| `gvm uninstall <version>` | Remove a specific installed version |
| `gvm default <version>` | Set the default Go version for new shells |

### Quick Example

```bash
# Install the latest stable version
gvm install latest

# Install a specific version
gvm install 1.21.5

# Switch to it
gvm use 1.21.5

# Verify
go version # outputs: go version go1.21.5 ...

```

---

## 📂 How it Works

GVM manages your `$GOROOT` and `$PATH` dynamically. It stores Go distributions in a directory (usually `/usr/local/gvm-versions`) and setups the gvm versions as required.

## 🤝 Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.

---

**Built with ❤️ for the Go community.**
