package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Folder struct {
	Path  string `yaml:"path"`
	Alias string `yaml:"alias"`
	Host  string `yaml:"host"`
}

type Config struct {
	DefaultHost string   `yaml:"ssh_host"`
	VSCodeExec  string   `yaml:"vscode_exec"`
	Folders     []Folder `yaml:"folders"`
}

func fail(msg string, err error) error {
	if err != nil {
		return fmt.Errorf("scode: %s: %w", msg, err)
	}
	return fmt.Errorf("scode: %s", msg)
}

func configPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "scode", "config.yaml")
}

func configTemplate() string {
	return strings.TrimSpace(`
ssh_host: prod
vscode_exec: code
folders:
  - path: /srv/projects/api
    alias: api
    host: prod
  - path: /home/user/infra
`) + "\n"
}

func missingConfigError(path string) error {
	return fmt.Errorf(
		"scode: no se encontró el archivo de configuración\n\nruta esperada:\n  %s\n\nestructura requerida:\n\n%s",
		path,
		configTemplate(),
	)
}

func folderAlias(f Folder) string {
	if f.Alias != "" {
		return f.Alias
	}
	return filepath.Base(f.Path)
}

func loadConfig() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, missingConfigError(path)
		}
		return nil, fail("error leyendo config", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fail("yaml inválido", err)
	}

	if cfg.VSCodeExec == "" {
		cfg.VSCodeExec = "code"
	}

	if _, err := exec.LookPath(cfg.VSCodeExec); err != nil {
		return nil, fail("no se encontró vscode_exec "+cfg.VSCodeExec, nil)
	}

	if cfg.DefaultHost == "" {
		return nil, fail("ssh_host es requerido", nil)
	}

	seen := map[string]bool{}
	for _, f := range cfg.Folders {
		if strings.TrimSpace(f.Path) == "" {
			return nil, fail("folder.path vacío", nil)
		}
		a := folderAlias(f)
		if seen[a] {
			return nil, fail("alias duplicado "+a, nil)
		}
		seen[a] = true
	}

	return &cfg, nil
}

func main() {
	root := &cobra.Command{
		Use:  "scode [alias]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			sort.Slice(cfg.Folders, func(i, j int) bool {
				return folderAlias(cfg.Folders[i]) < folderAlias(cfg.Folders[j])
			})

			if len(args) == 0 {
				for _, f := range cfg.Folders {
					fmt.Printf("%-12s %s\n", folderAlias(f), f.Path)
				}
				return nil
			}

			var sel *Folder
			for _, f := range cfg.Folders {
				if folderAlias(f) == args[0] {
					sel = &f
					break
				}
			}
			if sel == nil {
				return fail("alias no encontrado "+args[0], nil)
			}

			host := sel.Host
			if host == "" {
				host = cfg.DefaultHost
			}

			c := exec.Command(
				cfg.VSCodeExec,
				"--remote", "ssh-remote+"+host,
				sel.Path,
			)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin

			return c.Run()
		},
	}

	root.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := loadConfig()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var out []string
		for _, f := range cfg.Folders {
			a := folderAlias(f)
			if strings.HasPrefix(a, toComplete) {
				out = append(out, a)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}

	initCmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := configPath()
			if _, err := os.Stat(path); err == nil {
				return fail("ya existe "+path, nil)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fail("no se pudo crear directorio", err)
			}
			return os.WriteFile(path, []byte(configTemplate()), 0644)
		},
	}

	editCmd := &cobra.Command{
		Use: "edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			c := exec.Command(cfg.VSCodeExec, configPath())
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin
			return c.Run()
		},
	}

	root.AddCommand(initCmd, editCmd)
	root.Execute()
}
