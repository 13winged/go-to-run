package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config представляет основную конфигурацию утилиты
type Config struct {
	System   SystemConfig   `json:"system"`
	Security SecurityConfig `json:"security"`
	Packages PackagesConfig `json:"packages"`
}

// SystemConfig содержит настройки системы
type SystemConfig struct {
	Timezone string `json:"timezone"`
	Hostname string `json:"hostname"`
	SwapSize string `json:"swap_size"`
	Language string `json:"language"`
	Locale   string `json:"locale"`
}

// SecurityConfig содержит настройки безопасности
type SecurityConfig struct {
	SSHPort    int      `json:"ssh_port"`
	OpenPorts  []int    `json:"open_ports"`
	AllowIPs   []string `json:"allow_ips"`
	EnableUFW  bool     `json:"enable_ufw"`
	EnableFail2ban bool `json:"enable_fail2ban"`
	FirewallRules []FirewallRule `json:"firewall_rules"`
}

// FirewallRule представляет правило фаервола
type FirewallRule struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"`
	Comment  string `json:"comment"`
}

// PackagesConfig содержит настройки пакетов
type PackagesConfig struct {
	Basic       []string `json:"basic"`
	Network     []string `json:"network"`
	Monitoring  []string `json:"monitoring"`
	Development []string `json:"development"`
	Archive     []string `json:"archive"`
	Security    []string `json:"security"`
	System      []string `json:"system"`
	Database    []string `json:"database"`
	Web         []string `json:"web"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		System: SystemConfig{
			Timezone: "Europe/Moscow",
			Hostname: "",
			SwapSize: "2G",
			Language: "ru_RU",
			Locale:   "ru_RU.UTF-8",
		},
		Security: SecurityConfig{
			SSHPort:    22,
			OpenPorts:  []int{80, 443},
			AllowIPs:   []string{"127.0.0.1"},
			EnableUFW:  true,
			EnableFail2ban: true,
			FirewallRules: []FirewallRule{
				{Port: 22, Protocol: "tcp", Action: "allow", Comment: "SSH access"},
				{Port: 80, Protocol: "tcp", Action: "allow", Comment: "HTTP"},
				{Port: 443, Protocol: "tcp", Action: "allow", Comment: "HTTPS"},
			},
		},
		Packages: PackagesConfig{
			Basic: []string{
				"nano", "vim", "micro",
				"htop", "btop", "glances",
				"git", "curl", "wget", "rsync",
				"tree", "tmux", "screen", "zsh",
			},
			Archive: []string{
				"gzip", "gunzip", "zip", "unzip",
				"p7zip-full", "p7zip-rar", "unrar",
				"bzip2", "xz-utils", "zstd",
				"lz4", "tar", "cpio", "lzop",
			},
			Network: []string{
				"net-tools", "iproute2", "nmap",
				"traceroute", "mtr-tiny", "tcpdump",
				"openssh-client", "openssh-server",
				"dnsutils", "whois", "netcat-openbsd",
			},
			Monitoring: []string{
				"nmon", "iotop", "dstat", "vnstat",
				"atop", "sar", "sysstat",
			},
			Development: []string{
				"build-essential", "gcc", "g++",
				"python3", "python3-pip", "nodejs",
				"golang-go", "make", "cmake",
			},
			Security: []string{
				"ufw", "fail2ban", "rkhunter",
				"chkrootkit", "clamav",
			},
			System: []string{
				"mc", "ncdu", "bat", "fzf",
				"ripgrep", "jq", "yq",
			},
		},
	}
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурации: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации: %v", err)
	}

	return &config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(config *Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации конфигурации: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи конфигурации: %v", err)
	}

	return nil
}

// EnsureConfigDir создает директорию для конфигурации
func EnsureConfigDir() (string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "go-to-run")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории конфигурации: %v", err)
	}
	return configDir, nil
}

// GetConfigPath возвращает путь к конфигурационному файлу
func GetConfigPath() string {
	// 1. Текущая директория
	if _, err := os.Stat("go-to-run.json"); err == nil {
		return "go-to-run.json"
	}

	// 2. Пользовательская конфигурация
	configDir, err := EnsureConfigDir()
	if err == nil {
		userConfig := filepath.Join(configDir, "config.json")
		if _, err := os.Stat(userConfig); err == nil {
			return userConfig
		}
	}

	// 3. Глобальная конфигурация
	globalConfigs := []string{
		"/etc/go-to-run/config.json",
		"/usr/local/etc/go-to-run/config.json",
	}

	for _, config := range globalConfigs {
		if _, err := os.Stat(config); err == nil {
			return config
		}
	}

	// 4. Возвращаем путь для создания новой конфигурации
	return filepath.Join(configDir, "config.json")
}

// MergeConfigs объединяет две конфигурации
func MergeConfigs(base, override *Config) *Config {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base

	// Объединение настроек системы
	if override.System.Timezone != "" {
		merged.System.Timezone = override.System.Timezone
	}
	if override.System.Hostname != "" {
		merged.System.Hostname = override.System.Hostname
	}
	if override.System.SwapSize != "" {
		merged.System.SwapSize = override.System.SwapSize
	}

	// Объединение настроек безопасности
	if override.Security.SSHPort != 0 {
		merged.Security.SSHPort = override.Security.SSHPort
	}
	if len(override.Security.OpenPorts) > 0 {
		merged.Security.OpenPorts = override.Security.OpenPorts
	}
	if len(override.Security.AllowIPs) > 0 {
		merged.Security.AllowIPs = override.Security.AllowIPs
	}

	// Объединение пакетов
	mergePackageLists := func(base, override []string) []string {
		packageMap := make(map[string]bool)
		for _, pkg := range base {
			packageMap[pkg] = true
		}
		for _, pkg := range override {
			packageMap[pkg] = true
		}

		result := make([]string, 0, len(packageMap))
		for pkg := range packageMap {
			result = append(result, pkg)
		}
		return result
	}

	merged.Packages.Basic = mergePackageLists(merged.Packages.Basic, override.Packages.Basic)
	merged.Packages.Archive = mergePackageLists(merged.Packages.Archive, override.Packages.Archive)
	merged.Packages.Network = mergePackageLists(merged.Packages.Network, override.Packages.Network)
	merged.Packages.Monitoring = mergePackageLists(merged.Packages.Monitoring, override.Packages.Monitoring)
	merged.Packages.Development = mergePackageLists(merged.Packages.Development, override.Packages.Development)
	merged.Packages.Security = mergePackageLists(merged.Packages.Security, override.Packages.Security)
	merged.Packages.System = mergePackageLists(merged.Packages.System, override.Packages.System)

	return &merged
}

// ValidateConfig проверяет конфигурацию на корректность
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("конфигурация не может быть nil")
	}

	// Проверка часового пояса
	if config.System.Timezone == "" {
		return fmt.Errorf("часовой пояс не может быть пустым")
	}

	// Проверка портов
	for _, port := range config.Security.OpenPorts {
		if port < 1 || port > 65535 {
			return fmt.Errorf("некорректный порт: %d", port)
		}
	}

	// Проверка правил фаервола
	for _, rule := range config.Security.FirewallRules {
		if rule.Port < 1 || rule.Port > 65535 {
			return fmt.Errorf("некорректный порт в правиле: %d", rule.Port)
		}
		if rule.Protocol != "tcp" && rule.Protocol != "udp" {
			return fmt.Errorf("некорректный протокол в правиле: %s", rule.Protocol)
		}
		if rule.Action != "allow" && rule.Action != "deny" {
			return fmt.Errorf("некорректное действие в правиле: %s", rule.Action)
		}
	}

	return nil
}