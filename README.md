# scode

scode is a minimal Go CLI to open VS Code projects over SSH using short aliases defined in a YAML config file.

Features:
- Open remote projects by alias
- Shell autocompletion (bash/zsh/fish)
- YAML-based configuration
- Per-folder SSH host support
- Cross-platform (Linux, macOS, Windows)
- Uses existing SSH and VS Code setup (no credential handling)

Installation:
git clone <repo>
cd scode
make build
make install

Configuration:
scode init

Config location:
Linux: ~/.config/scode/config.yaml
macOS: ~/Library/Application Support/scode/config.yaml
Windows: %AppData%\scode\config.yaml

Example config.yaml:
ssh_host: prod
vscode_exec: code
folders:
  - path: /srv/projects/api
    alias: api
    host: prod
  - path: /home/user/infra

Usage:
scode        list projects
scode api    open project
scode edit   edit config


SSH:
Authentication is handled by ssh, ssh-agent, and VS Code Remote SSH.

