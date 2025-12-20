// Package dashboard предоставляет функциональность для отображения
// информационного дашборда системы (MOTD-style).
package dashboard

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/13winged/go-to-run/internal/config"
	"github.com/fatih/color"
)

// Dashboard управляет отображением информационной панели
type Dashboard struct {
	config *config.Config
}

// NewDashboard создает новый экземпляр дашборда
func NewDashboard() (*Dashboard, error) {
	// Загружаем конфигурацию (как это делает main.go)
	cfgPath := config.GetConfigPath()
	var cfg *config.Config

	if _, err := os.Stat(cfgPath); err == nil {
		cfg, _ = config.LoadConfig(cfgPath)
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &Dashboard{
		config: cfg,
	}, nil
}

// runCommand выполняет команду и возвращает вывод
func (d *Dashboard) runCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// runShell выполняет shell-команду
func (d *Dashboard) runShell(cmd string) (string, error) {
	return d.runCommand("sh", "-c", cmd)
}

// Render отображает дашборд в терминале
func (d *Dashboard) Render() error {
	d.renderHeader()
	d.renderSystemInfo()
	d.renderSecurityInfo()
	d.renderConfigInfo()
	d.renderUpdatesInfo()
	d.renderQuickActions()
	return nil
}

// renderHeader отображает заголовок дашборда
func (d *Dashboard) renderHeader() {
	blue := color.New(color.FgBlue, color.Bold)
	cyan := color.New(color.FgCyan)

	hostname, _ := os.Hostname()
	now := time.Now().Format("Monday, 02 January 2006 15:04:05 MST")

	fmt.Println()
	blue.Println("╔══════════════════════════════════════════════════════════════╗")
	blue.Println("║                 Go-to-Run System Dashboard                  ║")
	cyan.Printf("║    Host: %-45s    ║\n", hostname)
	cyan.Printf("║    Time: %-45s    ║\n", now)
	blue.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// renderSystemInfo отображает информацию о системе
func (d *Dashboard) renderSystemInfo() {
	green := color.New(color.FgGreen, color.Bold)
	green.Println("📊 SYSTEM INFORMATION")

	// Получаем системную информацию
	hostname, _ := os.Hostname()
	uptime, _ := d.runShell("uptime -p | sed 's/up //'")
	load, _ := d.runShell("cat /proc/loadavg | awk '{print $1, $2, $3}'")
	memory, _ := d.runShell("free -m | awk 'NR==2{printf \"%.1f/%.1fGB (%.0f%%)\", $3/1024,$2/1024,$3*100/$2 }'")
	osInfo, _ := d.runShell("grep PRETTY_NAME /etc/os-release 2>/dev/null | cut -d='\"' -f2 || echo 'Unknown'")
	kernel, _ := d.runCommand("uname", "-r")
	processes, _ := d.runShell("ps -e --no-headers | wc -l")

	fmt.Printf("├─ Hostname: %s\n", hostname)
	fmt.Printf("├─ OS: %s\n", osInfo)
	fmt.Printf("├─ Kernel: %s\n", kernel)
	if uptime != "" {
		fmt.Printf("├─ Uptime: %s\n", uptime)
	}
	if load != "" {
		fmt.Printf("├─ Load: %s\n", load)
	}
	if memory != "" {
		fmt.Printf("├─ Memory: %s\n", memory)
	}
	if processes != "" {
		fmt.Printf("└─ Processes: %s\n", processes)
	}
	fmt.Println()
}

// renderSecurityInfo отображает информацию о безопасности
func (d *Dashboard) renderSecurityInfo() {
	magenta := color.New(color.FgMagenta, color.Bold)
	magenta.Println("🛡️  SECURITY STATUS")

	// SSH статус
	sshStatus, _ := d.runShell("systemctl is-active ssh 2>/dev/null || systemctl is-active sshd 2>/dev/null || echo 'unknown'")
	sshIcon := "✅"
	if sshStatus != "active" {
		sshIcon = "⚠️ "
	}
	fmt.Printf("├─ SSH: %s %s\n", sshIcon, sshStatus)

	// SSH порт
	sshPort := "22"
	if d.config != nil && d.config.Security.SSHPort != 0 {
		sshPort = strconv.Itoa(d.config.Security.SSHPort)
	}
	fmt.Printf("├─ SSH Port: %s\n", sshPort)

	// UFW статус
	ufwStatus, _ := d.runShell("which ufw >/dev/null 2>&1 && ufw status | grep -q 'Status: active' && echo 'active' || echo 'inactive'")
	ufwIcon := "✅"
	if ufwStatus != "active" {
		ufwIcon = "❌"
	}
	fmt.Printf("├─ UFW: %s %s\n", ufwIcon, ufwStatus)

	// Fail2Ban статус
	fail2banStatus, _ := d.runShell("which fail2ban-client >/dev/null 2>&1 && fail2ban-client status 2>/dev/null | grep -q 'Status' && echo 'active' || echo 'not installed'")
	fail2banIcon := "✅"
	if fail2banStatus != "active" {
		fail2banIcon = "⚠️ "
	}
	fmt.Printf("└─ Fail2Ban: %s %s\n", fail2banIcon, fail2banStatus)

	fmt.Println()
}

// renderConfigInfo отображает информацию о конфигурации go-to-run
func (d *Dashboard) renderConfigInfo() {
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Println("⚙️  GO-TO-RUN CONFIGURATION")

	if d.config == nil {
		fmt.Println("   Using default configuration")
		fmt.Println()
		return
	}

	fmt.Printf("├─ Timezone: %s\n", d.config.System.Timezone)

	if d.config.System.Hostname != "" {
		fmt.Printf("├─ Hostname: %s\n", d.config.System.Hostname)
	}

	fmt.Printf("├─ Swap: %s\n", d.config.System.SwapSize)

	// Показываем разрешенные порты
	fmt.Printf("├─ Open Ports: ")
	if len(d.config.Security.OpenPorts) > 0 {
		for i, port := range d.config.Security.OpenPorts {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", port)
		}
		fmt.Println()
	} else {
		fmt.Println("none")
	}

	// Показываем IP-адреса
	fmt.Printf("├─ Allowed IPs: ")
	if len(d.config.Security.AllowIPs) > 0 {
		for i, ip := range d.config.Security.AllowIPs {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", ip)
		}
		fmt.Println()
	} else {
		fmt.Println("none")
	}

	// Показываем количество пакетов по категориям
	fmt.Println("└─ Package Categories:")
	categories := map[string][]string{
		"Basic":       d.config.Packages.Basic,
		"Network":     d.config.Packages.Network,
		"Development": d.config.Packages.Development,
		"Security":    d.config.Packages.Security,
		"System":      d.config.Packages.System,
		"Archive":     d.config.Packages.Archive,
		"Database":    d.config.Packages.Database,
		"Web":         d.config.Packages.Web,
	}

	for name, packages := range categories {
		if len(packages) > 0 {
			fmt.Printf("   • %s: %d packages\n", name, len(packages))
		}
	}

	fmt.Println()
}

// renderUpdatesInfo отображает информацию об обновлениях
func (d *Dashboard) renderUpdatesInfo() {
	yellow := color.New(color.FgYellow, color.Bold)
	yellow.Println("📦 AVAILABLE UPDATES")

	// Проверяем разные менеджеры пакетов
	updateCount := 0

	// APT (Debian/Ubuntu)
	if aptUpdates, err := d.runShell("which apt >/dev/null 2>&1 && apt list --upgradable 2>/dev/null | wc -l"); err == nil && aptUpdates != "" {
		if count, err := strconv.Atoi(aptUpdates); err == nil && count > 1 {
			updateCount = count - 1
			fmt.Printf("├─ APT: %d updates available\n", updateCount)
		}
	}

	// DNF (Fedora/RHEL)
	if dnfUpdates, err := d.runShell("which dnf >/dev/null 2>&1 && dnf check-update --quiet 2>/dev/null | wc -l"); err == nil && dnfUpdates != "" {
		if count, err := strconv.Atoi(dnfUpdates); err == nil && count > 0 {
			updateCount = count
			fmt.Printf("├─ DNF: %d updates available\n", updateCount)
		}
	}

	// YUM (CentOS/RHEL)
	if yumUpdates, err := d.runShell("which yum >/dev/null 2>&1 && yum check-update --quiet 2>/dev/null | wc -l"); err == nil && yumUpdates != "" {
		if count, err := strconv.Atoi(yumUpdates); err == nil && count > 0 {
			updateCount = count
			fmt.Printf("├─ YUM: %d updates available\n", updateCount)
		}
	}

	if updateCount == 0 {
		fmt.Println("├─ ✅ System is up to date")
	}

	// Время последнего обновления
	if lastUpdate, err := d.runShell("stat -c %y /var/lib/apt/periodic/update-success-stamp 2>/dev/null || echo 'Never'"); err == nil {
		if lastUpdate != "Never" {
			lastUpdateTime, err := time.Parse("2006-01-02 15:04:05.000000000 -0700", lastUpdate)
			if err == nil {
				fmt.Printf("└─ Last update: %s ago\n", time.Since(lastUpdateTime).Round(time.Hour))
			}
		} else {
			fmt.Println("└─ Last update: Never")
		}
	}

	fmt.Println()
}

// renderQuickActions отображает подсказки по быстрым действиям
func (d *Dashboard) renderQuickActions() {
	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("🚀 QUICK ACTIONS")

	fmt.Println("   sudo go-to-run --update           Update system packages")
	fmt.Println("   sudo go-to-run --install          Install configured packages")
	fmt.Println("   sudo go-to-run --security         Configure security")
	fmt.Println("   sudo go-to-run --clean            Clean system")
	fmt.Println("   go-to-run --info                  Show detailed system info")
	fmt.Println()
	fmt.Println("   go-to-run check                   Check system status")
	fmt.Println("   go-to-run monitor                 Real-time monitoring")
	fmt.Println("   go-to-run backup                  Backup configuration")
	fmt.Println()
}
