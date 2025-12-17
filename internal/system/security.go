package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/briandowns/spinner"
)

// SecurityManager управляет настройками безопасности
type SecurityManager struct{}

// FirewallConfig содержит настройки фаервола
type FirewallConfig struct {
	Enabled    bool
	SSHPort    int
	OpenPorts  []int
	AllowIPs   []string
	Rules      []FirewallRule
}

// FirewallRule представляет правило фаервола
type FirewallRule struct {
	Port     int
	Protocol string
	Action   string
	Comment  string
}

// SetupFirewall настраивает фаервол
func (sm *SecurityManager) SetupFirewall(config *FirewallConfig) error {
	if !config.Enabled {
		fmt.Println("Настройка фаервола отключена в конфигурации")
		return nil
	}

	// Проверяем установлен ли UFW
	if !sm.isUFWInstalled() {
		fmt.Println("UFW не установлен, устанавливаем...")
		if err := sm.installUFW(); err != nil {
			return fmt.Errorf("ошибка установки UFW: %v", err)
		}
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Настройка фаервола..."
	s.Start()
	defer s.Stop()

	// Проверяем статус UFW
	status, err := sm.getUFWStatus()
	if err != nil {
		return fmt.Errorf("ошибка получения статуса UFW: %v", err)
	}

	// Если фаервол уже активен, показываем правила
	if strings.Contains(status, "Status: active") {
		fmt.Println("UFW уже активен")
		sm.showUFWRules()
		return nil
	}

	// Сбрасываем правила если фаервол отключен
	if strings.Contains(status, "Status: inactive") {
		if err := sm.resetUFW(); err != nil {
			return fmt.Errorf("ошибка сброса UFW: %v", err)
		}

		// Настраиваем политики по умолчанию
		if err := sm.setDefaultPolicies(); err != nil {
			return fmt.Errorf("ошибка настройки политик: %v", err)
		}

		// Применяем правила
		if err := sm.applyRules(config); err != nil {
			return fmt.Errorf("ошибка применения правил: %v", err)
		}

		// Включаем логирование
		if err := sm.enableLogging(); err != nil {
			return fmt.Errorf("ошибка включения логирования: %v", err)
		}

		// Включаем фаервол
		if err := sm.enableUFW(); err != nil {
			return fmt.Errorf("ошибка включения UFW: %v", err)
		}
	}

	fmt.Println("Фаервол успешно настроен")
	sm.showUFWStatus()
	return nil
}

// SetupFail2ban настраивает Fail2ban
func (sm *SecurityManager) SetupFail2ban() error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Настройка Fail2ban..."
	s.Start()
	defer s.Stop()

	// Проверяем установлен ли Fail2ban
	if !sm.isFail2banInstalled() {
		if err := sm.installFail2ban(); err != nil {
			return fmt.Errorf("ошибка установки Fail2ban: %v", err)
		}
	}

	// Создаем конфигурацию
	if err := sm.createFail2banConfig(); err != nil {
		return fmt.Errorf("ошибка создания конфигурации Fail2ban: %v", err)
	}

	// Перезапускаем службу
	if err := sm.restartFail2ban(); err != nil {
		return fmt.Errorf("ошибка перезапуска Fail2ban: %v", err)
	}

	fmt.Println("Fail2ban успешно настроен")
	return nil
}

// SetupSSH настраивает SSH
func (sm *SecurityManager) SetupSSH(port int, allowRoot bool, passwordAuth bool) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Настройка SSH..."
	s.Start()
	defer s.Stop()

	// Создаем резервную копию конфигурации
	if err := sm.backupSSHConfig(); err != nil {
		return fmt.Errorf("ошибка создания бэкапа SSH: %v", err)
	}

	// Настраиваем SSH
	if err := sm.configureSSH(port, allowRoot, passwordAuth); err != nil {
		return fmt.Errorf("ошибка настройки SSH: %v", err)
	}

	// Перезапускаем службу SSH
	if err := sm.restartSSH(); err != nil {
		return fmt.Errorf("ошибка перезапуска SSH: %v", err)
	}

	fmt.Println("SSH успешно настроен")
	return nil
}

// Helper методы

func (sm *SecurityManager) isUFWInstalled() bool {
	_, err := exec.LookPath("ufw")
	return err == nil
}

func (sm *SecurityManager) installUFW() error {
	pm, err := (&PackageManagerDetector{}).Detect()
	if err != nil {
		return err
	}
	return exec.Command("sh", "-c", pm.Install+" ufw").Run()
}

func (sm *SecurityManager) getUFWStatus() (string, error) {
	output, err := exec.Command("ufw", "status").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (sm *SecurityManager) resetUFW() error {
	return exec.Command("ufw", "--force", "reset").Run()
}

func (sm *SecurityManager) setDefaultPolicies() error {
	// Отключаем входящие соединения по умолчанию
	if err := exec.Command("ufw", "default", "deny", "incoming").Run(); err != nil {
		return err
	}
	// Разрешаем исходящие соединения по умолчанию
	return exec.Command("ufw", "default", "allow", "outgoing").Run()
}

func (sm *SecurityManager) applyRules(config *FirewallConfig) error {
	seenPorts := make(map[int]bool)

	// Добавляем SSH порт
	if config.SSHPort > 0 {
		if err := sm.addPortRule(config.SSHPort, "tcp", "SSH access"); err != nil {
			return err
		}
		seenPorts[config.SSHPort] = true
	}

	// Добавляем другие порты
	for _, port := range config.OpenPorts {
		if port <= 0 || port > 65535 || seenPorts[port] {
			continue
		}
		if err := sm.addPortRule(port, "tcp", fmt.Sprintf("Port %d", port)); err != nil {
			return err
		}
		seenPorts[port] = true
	}

	// Добавляем пользовательские правила
	for _, rule := range config.Rules {
		if err := sm.addCustomRule(rule); err != nil {
			return err
		}
	}

	// Разрешаем указанные IP-адреса
	for _, ip := range config.AllowIPs {
		if err := sm.allowIP(ip); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SecurityManager) addPortRule(port int, protocol, comment string) error {
	cmd := fmt.Sprintf("ufw allow %d/%s comment '%s'", port, protocol, comment)
	return exec.Command("sh", "-c", cmd).Run()
}

func (sm *SecurityManager) addCustomRule(rule FirewallRule) error {
	var cmd string
	switch rule.Action {
	case "allow":
		cmd = fmt.Sprintf("ufw allow %d/%s", rule.Port, rule.Protocol)
	case "deny":
		cmd = fmt.Sprintf("ufw deny %d/%s", rule.Port, rule.Protocol)
	default:
		return fmt.Errorf("неподдерживаемое действие: %s", rule.Action)
	}

	if rule.Comment != "" {
		cmd += fmt.Sprintf(" comment '%s'", rule.Comment)
	}

	return exec.Command("sh", "-c", cmd).Run()
}

func (sm *SecurityManager) allowIP(ip string) error {
	return exec.Command("ufw", "allow", "from", ip).Run()
}

func (sm *SecurityManager) enableLogging() error {
	return exec.Command("ufw", "logging", "on").Run()
}

func (sm *SecurityManager) enableUFW() error {
	return exec.Command("sh", "-c", "yes | ufw enable").Run()
}

func (sm *SecurityManager) showUFWStatus() {
	output, err := exec.Command("ufw", "status", "verbose").Output()
	if err == nil {
		fmt.Println(string(output))
	}
}

func (sm *SecurityManager) showUFWRules() {
	output, err := exec.Command("ufw", "status", "numbered").Output()
	if err == nil {
		fmt.Println(string(output))
	}
}

func (sm *SecurityManager) isFail2banInstalled() bool {
	_, err := exec.LookPath("fail2ban-client")
	return err == nil
}

func (sm *SecurityManager) installFail2ban() error {
	pm, err := (&PackageManagerDetector{}).Detect()
	if err != nil {
		return err
	}
	return exec.Command("sh", "-c", pm.Install+" fail2ban").Run()
}

func (sm *SecurityManager) createFail2banConfig() error {
	config := `[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5
ignoreip = 127.0.0.1/8

[sshd]
enabled = true
port = ssh
logpath = %(sshd_log)s
backend = %(sshd_backend)s
`

	configPath := "/etc/fail2ban/jail.local"
	return os.WriteFile(configPath, []byte(config), 0644)
}

func (sm *SecurityManager) restartFail2ban() error {
	// Включаем автозагрузку
	if err := exec.Command("systemctl", "enable", "fail2ban").Run(); err != nil {
		return err
	}
	// Перезапускаем службу
	return exec.Command("systemctl", "restart", "fail2ban").Run()
}

func (sm *SecurityManager) backupSSHConfig() error {
	backupCmd := "cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d%H%M%S)"
	return exec.Command("sh", "-c", backupCmd).Run()
}

func (sm *SecurityManager) configureSSH(port int, allowRoot, passwordAuth bool) error {
	configPath := "/etc/ssh/sshd_config"
	config, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(config), "\n")
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Пропускаем комментарии
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			newLines = append(newLines, line)
			continue
		}

		// Изменяем настройки
		switch {
		case strings.HasPrefix(trimmed, "Port "):
			newLines = append(newLines, fmt.Sprintf("Port %d", port))
		case strings.HasPrefix(trimmed, "PermitRootLogin "):
			value := "no"
			if allowRoot {
				value = "yes"
			}
			newLines = append(newLines, fmt.Sprintf("PermitRootLogin %s", value))
		case strings.HasPrefix(trimmed, "PasswordAuthentication "):
			value := "no"
			if passwordAuth {
				value = "yes"
			}
			newLines = append(newLines, fmt.Sprintf("PasswordAuthentication %s", value))
		default:
			newLines = append(newLines, line)
		}
	}

	// Добавляем рекомендуемые настройки
	recommendedSettings := []string{
		"",
		"# Additional security settings",
		"Protocol 2",
		"ClientAliveInterval 300",
		"ClientAliveCountMax 2",
		"MaxAuthTries 3",
		"MaxSessions 10",
		"X11Forwarding no",
	}

	newLines = append(newLines, recommendedSettings...)

	return os.WriteFile(configPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func (sm *SecurityManager) restartSSH() error {
	return exec.Command("systemctl", "restart", "ssh").Run()
}

// CheckSecurity проверяет безопасность системы
func (sm *SecurityManager) CheckSecurity() error {
	fmt.Println("Проверка безопасности системы...")

	// Проверяем открытые порты
	fmt.Println("\n1. Проверка открытых портов:")
	if err := sm.checkOpenPorts(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	}

	// Проверяем обновления безопасности
	fmt.Println("\n2. Проверка обновлений безопасности:")
	if err := sm.checkSecurityUpdates(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	}

	// Проверяем UFW
	fmt.Println("\n3. Проверка фаервола:")
	sm.checkUFW()

	// Проверяем Fail2ban
	fmt.Println("\n4. Проверка Fail2ban:")
	sm.checkFail2ban()

	return nil
}

func (sm *SecurityManager) checkOpenPorts() error {
	cmd := "ss -tulpn | grep LISTEN"
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	fmt.Printf("Найдено %d открытых портов:\n", len(lines)-1)
	for _, line := range lines {
		if line != "" {
			fmt.Printf("  %s\n", line)
		}
	}

	return nil
}

func (sm *SecurityManager) checkSecurityUpdates() error {
	pm, err := (&PackageManagerDetector{}).Detect()
	if err != nil {
		return err
	}

	updates, err := GetAvailableUpdates(pm)
	if err != nil {
		return err
	}

	fmt.Printf("Доступно %d обновлений\n", len(updates))
	if len(updates) > 0 {
		fmt.Println("Рекомендуемые обновления безопасности:")
		for i, update := range updates {
			if i < 10 { // Показываем только первые 10
				fmt.Printf("  %s\n", update)
			}
		}
	}

	return nil
}

func (sm *SecurityManager) checkUFW() {
	if sm.isUFWInstalled() {
		status, err := sm.getUFWStatus()
		if err == nil {
			fmt.Printf("UFW статус: %s", status)
		}
	} else {
		fmt.Println("UFW не установлен")
	}
}

func (sm *SecurityManager) checkFail2ban() {
	if sm.isFail2banInstalled() {
		output, err := exec.Command("fail2ban-client", "status").Output()
		if err == nil {
			fmt.Printf("Fail2ban статус:\n%s", string(output))
		}
	} else {
		fmt.Println("Fail2ban не установлен")
	}
}