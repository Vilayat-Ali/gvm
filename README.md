# 🐹 GVM: Go Version Manager

<div align="center">

![Version](https://img.shields.io/badge/version-3.0.0-blue?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)

**Install and switch between Go toolchains in milliseconds — without root, and without touching a single system file.**

[Install](#-installation) • [Quick start](#-quick-start) • [Commands](#-commands) • [How it works](#-how-it-works) • [Safety](#-safety) • [FAQ](#-faq)

</div>

---

## ✨ Features

- **No root required.** Everything lives under `~/.local/share/gvm`. gvm refuses to run under `sudo` and never asks for elevation.
- **Never touches system files.** `/usr/local/go`, `/usr`, `/etc` and friends are on a hard-coded deny list. Your distro's Go stays exactly where it is.
- **Instant switching.** Changing versions moves one symlink — about 10 ms, no re-extraction.
- **Verified downloads.** Every archive is checked against the SHA-256 published by go.dev, hashed while it streams to disk.
- **Atomic and crash-safe.** Downloads, installs, config writes and version switches all land via `rename(2)`. An interrupted command never leaves you without a working Go.
- **Hardened extraction.** Path traversal, absolute paths, escaping symlinks and device nodes are all rejected.
- **`gvm doctor`.** Explains PATH problems, shadowing, a pinned `GOROOT`, and Go's automatic toolchain switching — then tells you how to fix each one.

---

## 🚀 Installation

### Install script

```bash
curl -fsSL https://raw.githubusercontent.com/Vilayat-Ali/gvm/main/scripts/install.sh | bash
```

It verifies the release checksum, installs to `~/.local/bin`, and adds gvm to your shell profile. It refuses to run as root.

### From source

```bash
git clone https://github.com/Vilayat-Ali/gvm.git
cd gvm
make install          # builds and installs to ~/.local/bin
```

### Manual

Download the archive for your platform from the [Releases page](https://github.com/Vilayat-Ali/gvm/releases), then:

```bash
tar -xzf gvm-linux-amd64-<version>.tar.gz
install -m 0755 gvm ~/.local/bin/gvm
```

Add both directories to your PATH:

```bash
export PATH="$HOME/.local/share/gvm/bin:$HOME/.local/bin:$PATH"
```

---

## 🎯 Quick start

```bash
gvm configure       # create the gvm directory and fetch the release catalog
gvm use latest      # install and activate the newest stable Go
gvm doctor          # confirm your shell is wired up correctly

go version
# go version go1.27.0 linux/amd64
```

Switching afterwards is instant:

```bash
gvm use 1.24        # resolves to the newest 1.24.x
gvm use 1.23.4
```

---

## 📚 Commands

| Command | What it does |
|---|---|
| `gvm configure` | Create the gvm directory and fetch the version catalog |
| `gvm list` | Show installed versions (`*` marks the active one) |
| `gvm list --remote` | Show versions available to download |
| `gvm list --current` | Print only the active version |
| `gvm list update` | Refresh the catalog from go.dev |
| `gvm download <version>` | Download and unpack a version without activating it |
| `gvm use <version>` | Activate a version, installing it first if needed |
| `gvm remove <version>` | Remove an installed version (`--purge` also drops the cached archive) |
| `gvm doctor` | Diagnose PATH, shadowing, `GOROOT` and toolchain-switching problems |
| `gvm env` | Print the shell line that puts gvm on your PATH |

### Version arguments

All commands accept the same forms:

```
1.24.2    v1.24.2    go1.24.2      exact version
1.24                                newest 1.24.x
1.25rc1                             release candidates and betas
latest                              newest stable release
```

### Examples

```console
$ gvm list

  Installed
  ---------
  * 1.27.0
    1.24.13
    1.23.12

  * = active   |   `gvm list --remote` shows downloadable versions
```

```console
$ gvm doctor

  Environment
  -----------
    root         /home/you/.local/share/gvm
    shims        /home/you/.local/share/gvm/bin
    installed    3 version(s)
    active       1.27.0
    go on PATH   /home/you/.local/share/gvm/bin/go (1.27.0)

  everything looks good
```

---

## 🔍 How it works

```
~/.local/share/gvm/
├── bin/
│   ├── go       -> ../current/bin/go
│   └── gofmt    -> ../current/bin/gofmt
├── cache/       downloaded, checksum-verified archives
├── current      -> versions/go1.27.0
└── versions/
    ├── go1.27.0/
    └── go1.24.13/
```

`~/.local/share/gvm/bin` is the only thing on your PATH. Switching versions repoints `current` with an atomic rename, so `go` resolves to the new toolchain immediately — no files are copied, deleted or re-extracted.

Installing a version stages the extraction in a temporary directory, runs `go version` to prove the toolchain works, and only then renames it into place. If anything fails, the previous state is untouched.

The release catalog and every checksum come from the official `https://go.dev/dl/?mode=json` index.

Override the locations with `GVM_ROOT` and `XDG_CONFIG_HOME` if you need to; `GVM_API_URL` and `GVM_DOWNLOAD_URL` let you point at a mirror.

---

## 🛡 Safety

gvm is designed so that a bug, a hostile archive or a mistyped argument cannot damage your system:

- **Deny list.** `/`, `/usr`, `/usr/local`, `/usr/local/go`, `/etc`, `/var`, `/home`, your home directory itself and ~25 others can never be used as the gvm root or be deleted.
- **Confined deletes.** Every removal is checked to be inside the gvm root first, and symlinks are unlinked rather than followed — so a stray `current` link can never delete what it points at.
- **Strict version parsing.** Version arguments must match `go<major>.<minor>[.<patch>][rc|beta|alpha<n>]`, so nothing that reaches the filesystem can contain `..`, `/` or shell metacharacters.
- **Extraction guards.** Entries with absolute paths, `..` components, symlinks pointing outside the destination, device nodes and FIFOs are skipped; oversized archives are rejected.
- **No `tar` subprocess, no shell.** Archives are unpacked in-process with `archive/tar`.
- **No setuid, no sudo.** Running `sudo gvm` is refused outright, because it would leave root-owned files in your home directory.
- **Refuses to remove the active version** until you switch away from it.

`gvm` does not modify `/usr/local/go`. If a system Go is shadowing gvm on your PATH, `gvm doctor` reports it and tells you how to reorder your PATH — it will not delete it for you.

---

## 🔧 Development

```bash
make check          # verify-go + fmt-check + vet + test
make test-race      # tests under the race detector
make test-coverage  # writes coverage.html
make build          # build for the current platform
make release        # cross-compile and write SHA256SUMS.txt
make help           # all targets
```

### Changing the Go version used to build gvm

The required toolchain is declared **once**, in `go.mod`:

```
go 1.25.5
```

The Makefile reads it (`make verify-go` checks your local toolchain against it) and CI installs it via `go-version-file: go.mod`. Edit `go.mod`, run `make deps`, and everything follows.

---

## ❓ FAQ

**Do I need sudo?**
No. gvm installs into your home directory. Running it with `sudo` is refused unless you set `GVM_ALLOW_ROOT=1`.

**Will it delete my existing Go installation?**
No. gvm never writes to or deletes anything outside its own root directory. If your distro's Go at `/usr/local/go` comes first on your PATH, `gvm doctor` will point that out so you can reorder it.

**I ran `gvm use 1.23`, but `go version` reports something else.**
Go switches toolchains on its own. If the `go.mod` in your current directory requires a newer Go than the one you activated, Go downloads that version and runs it instead — this is Go's default `GOTOOLCHAIN=auto` behaviour, not gvm. `gvm doctor` detects this and names the `go.mod` responsible. Either `gvm use` the version the module wants, or set `GOTOOLCHAIN=local` to make Go report an error instead of switching.

**How fast is switching?**
About 10 ms for an already-installed version — it is a single symlink rename.

**Are downloads verified?**
Yes, against the SHA-256 published by go.dev. The hash is computed while the file streams to disk, and a mismatch deletes the partial download.

**Which platforms are supported?**
Linux and macOS, on amd64 and arm64.

**Where does everything live?**
Toolchains in `~/.local/share/gvm` (override with `GVM_ROOT`), config in `~/.config/gvm/config.json`.

**How do I remove gvm completely?**

```bash
make clean-setup                      # or:
rm -rf ~/.local/share/gvm ~/.config/gvm ~/.local/bin/gvm
```

Then remove the `# added by gvm installer` line from your shell profile.

---

## 🤝 Contributing

1. Fork and clone the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and run `make check`
4. Commit, push, and open a Pull Request

---

## 📄 License

MIT — see [LICENSE](LICENSE).

---

## 🙏 Acknowledgments

- Inspired by [nvm](https://github.com/nvm-sh/nvm) and [fnm](https://github.com/Schniz/fnm)
- Powered by [Cobra](https://github.com/spf13/cobra)
- Go downloads served by [go.dev](https://go.dev/)

---

<div align="center">

**Built with 💜 for the Go community**

</div>
