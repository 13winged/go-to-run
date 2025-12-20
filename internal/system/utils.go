package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// SystemInfo содержит информацию о системе
type SystemInfo struct {
	Distro      string
	Version     string
	Kernel      string
	Uptime      string
	Memory      string
	Disk        string
	CPU         string
	IPAddress   string
	Processes   int
	LoadAverage string
}

// SystemUtils предоставляет утилиты для работы с системой
type SystemUtils struct{}

// GetSystemInfo собирает информацию о системе
func (su *SystemUtils) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// Определяем дистрибутив
	if distro, version, err := detectDistro(); err == nil {
		info.Distro = distro
		info.Version = version
	}

	// Получаем информацию о ядре
	if kernel, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(kernel))
	}

	// Получаем время работы
	if uptime, err := exec.Command("uptime", "-p").Output(); err == nil {
		info.Uptime = strings.TrimSpace(strings.TrimPrefix(string(uptime), "up "))
	}

	// Получаем информацию о памяти
	if memory, err := exec.Command("free", "-h").Output(); err == nil {
		lines := strings.Split(string(memory), "\n")
		if len(lines) > 1 {
			parts := strings.Fields(lines[1])
			if len(parts) >= 7 {
				info.Memory = fmt.Sprintf("Total: %s, Used: %s, Free: %s", parts[1], parts[2], parts[6])
			}
		}
	}

	// Получаем информацию о дисках
	if disk, err := exec.Command("df", "-h", "--output=source,size,used,avail,pcent,target").Output(); err == nil {
		lines := strings.Split(string(disk), "\n")
		var diskInfo []string
		for i, line := range lines {
			if i > 0 && len(line) > 0 {
				diskInfo = append(diskInfo, line)
			}
		}
		if len(diskInfo) > 0 {
			info.Disk = strings.Join(diskInfo[:min(3, len(diskInfo))], "; ")
		}
	}

	// Получаем информацию о CPU
	if cpu, err := exec.Command("lscpu").Output(); err == nil {
		lines := strings.Split(string(cpu), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Model name:") {
				info.CPU = strings.TrimSpace(strings.Split(line, ":")[1])
				break
			}
		}
	}

	// Получаем IP адрес
	if ip, err := exec.Command("hostname", "-I").Output(); err == nil {
		info.IPAddress = strings.TrimSpace(string(ip))
	}

	// Получаем количество процессов
	if procs, err := exec.Command("ps", "-e", "--no-headers").Output(); err == nil {
		info.Processes = len(strings.Split(strings.TrimSpace(string(procs)), "\n"))
	}

	// Получаем среднюю загрузку
	if load, err := exec.Command("uptime").Output(); err == nil {
		parts := strings.Split(string(load), "load average:")
		if len(parts) > 1 {
			info.LoadAverage = strings.TrimSpace(parts[1])
		}
	}

	return info, nil
}

// SetupTimezone настраивает часовой пояс
func (su *SystemUtils) SetupTimezone(timezone string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Настройка часового пояса: %s", timezone)
	s.Start()
	defer s.Stop()

	if commandExists("timedatectl") {
		if err := exec.Command("timedatectl", "set-timezone", timezone).Run(); err != nil {
			// Альтернативный метод
			return su.setTimezoneFile(timezone)
		}
		return nil
	}
	return su.setTimezoneFile(timezone)
}

func (su *SystemUtils) setTimezoneFile(timezone string) error {
	// Проверяем существование часового пояса
	zoneInfo := fmt.Sprintf("/usr/share/zoneinfo/%s", timezone)
	if _, err := os.Stat(zoneInfo); err != nil {
		return fmt.Errorf("часовой пояс не найден: %s", timezone)
	}

	// Удаляем старый симлинк
	os.Remove("/etc/localtime")

	// Создаем новый симлинк
	if err := os.Symlink(zoneInfo, "/etc/localtime"); err != nil {
		return fmt.Errorf("ошибка создания симлинка: %v", err)
	}

	// Записываем в /etc/timezone
	return os.WriteFile("/etc/timezone", []byte(timezone+"\n"), 0644)
}

// SetupLocale настраивает локаль
func (su *SystemUtils) SetupLocale(locale string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Настройка локали: %s", locale)
	s.Start()
	defer s.Stop()

	if !commandExists("locale-gen") {
		return fmt.Errorf("locale-gen не найден")
	}

	// Генерируем локаль
	cmd := fmt.Sprintf("locale-gen %s", locale)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		return fmt.Errorf("ошибка генерации локали: %v", err)
	}

	// Обновляем настройки локали
	cmd = fmt.Sprintf("update-locale LANG=%s LC_ALL=%s", locale, locale)
	return exec.Command("sh", "-c", cmd).Run()
}

// SetupSwap настраивает swap
func (su *SystemUtils) SetupSwap(swapSize string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Настройка swap..."
	s.Start()
	defer s.Stop()

	// Проверяем существующий swap
	if swapInfo, err := exec.Command("swapon", "--show").Output(); err == nil {
		if strings.TrimSpace(string(swapInfo)) != "" {
			return fmt.Errorf("swap уже настроен")
		}
	}

	// Определяем размер если не указан
	if swapSize == "" {
		var err error
		swapSize, err = su.calculateSwapSize()
		if err != nil {
			return fmt.Errorf("ошибка расчета размера swap: %v", err)
		}
	}

	// Создаем swap файл
	swapFile := "/swapfile"
	if err := su.createSwapFile(swapFile, swapSize); err != nil {
		return err
	}

	// Настраиваем swap
	if err := su.configureSwap(swapFile); err != nil {
		return err
	}

	// Настраиваем swappiness
	return su.configureSwappiness()
}

func (su *SystemUtils) calculateSwapSize() (string, error) {
	memInfo, err := exec.Command("free", "-b").Output()
	if err != nil {
		return "2G", nil // Значение по умолчанию
	}

	lines := strings.Split(string(memInfo), "\n")
	if len(lines) > 1 {
		parts := strings.Fields(lines[1])
		if len(parts) >= 2 {
			memBytes, err := parseBytes(parts[1])
			if err != nil {
				return "2G", nil
			}

			// Рекомендуемый размер swap:
			// - RAM < 2GB: 2x RAM
			// - RAM 2-8GB: 1x RAM
			// - RAM > 8GB: 0.5x RAM
			var swapBytes uint64
			if memBytes < 2*1024*1024*1024 {
				swapBytes = memBytes * 2
			} else if memBytes <= 8*1024*1024*1024 {
				swapBytes = memBytes
			} else {
				swapBytes = memBytes / 2
			}

			return fmt.Sprintf("%dM", swapBytes/(1024*1024)), nil
		}
	}

	return "2G", nil
}

func (su *SystemUtils) createSwapFile(swapFile, size string) error {
	// Удаляем старый файл если существует
	os.Remove(swapFile)

	// Создаем файл с помощью fallocate
	cmd := fmt.Sprintf("fallocate -l %s %s", size, swapFile)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		// fallocate может не работать, используем dd
		cmd = fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%s status=progress",
			swapFile, strings.TrimSuffix(size, "G"))
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return fmt.Errorf("ошибка создания swap файла: %v", err)
		}
	}

	// Устанавливаем права
	return exec.Command("chmod", "600", swapFile).Run()
}

func (su *SystemUtils) configureSwap(swapFile string) error {
	// Форматируем как swap
	if err := exec.Command("mkswap", swapFile).Run(); err != nil {
		return fmt.Errorf("ошибка форматирования swap: %v", err)
	}

	// Включаем swap
	if err := exec.Command("swapon", swapFile).Run(); err != nil {
		return fmt.Errorf("ошибка включения swap: %v", err)
	}

	// Добавляем в fstab
	fstabEntry := fmt.Sprintf("%s none swap sw 0 0\n", swapFile)
	f, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия fstab: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(fstabEntry); err != nil {
		return fmt.Errorf("ошибка записи в fstab: %v", err)
	}

	return nil
}

func (su *SystemUtils) configureSwappiness() error {
	config := "vm.swappiness=10\nvm.vfs_cache_pressure=50\n"
	configFile := "/etc/sysctl.d/99-swappiness.conf"

	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("ошибка записи конфигурации swappiness: %v", err)
	}

	return exec.Command("sysctl", "-p", configFile).Run()
}

// CleanSystem очищает систему
func (su *SystemUtils) CleanSystem() error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Очистка системы..."
	s.Start()
	defer s.Stop()

	// Очищаем временные файлы
	su.cleanTempFiles()

	// Очищаем кеш пакетов
	su.cleanPackageCache()

	// Очищаем логи
	su.cleanLogs()

	// Очищаем кеш systemd
	su.cleanSystemdCache()

	return nil
}

func (su *SystemUtils) cleanTempFiles() {
	exec.Command("sh", "-c", "rm -rf /tmp/* 2>/dev/null || true").Run()
	exec.Command("sh", "-c", "rm -rf /var/tmp/* 2>/dev/null || true").Run()
}

func (su *SystemUtils) cleanPackageCache() {
	pm, err := (&PackageManagerDetector{}).Detect()
	if err == nil {
		exec.Command("sh", "-c", pm.Clean).Run()
	}
}

func (su *SystemUtils) cleanLogs() {
	exec.Command("sh", "-c", "find /var/log -type f -name '*.gz' -delete 2>/dev/null || true").Run()
	exec.Command("sh", "-c", "find /var/log -type f -name '*.1' -delete 2>/dev/null || true").Run()
}

func (su *SystemUtils) cleanSystemdCache() {
	if commandExists("journalctl") {
		exec.Command("sh", "-c", "journalctl --vacuum-time=3d").Run()
	}
}

// RunCommand выполняет команду с выводом
func (su *SystemUtils) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunCommandOutput выполняет команду и возвращает вывод
func (su *SystemUtils) RunCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}
		return string(output), fmt.Errorf("%w: %s", err, string(stderr))
	}
	return string(output), err
}

// Helper функции

func detectDistro() (string, string, error) {
	if _, err := os.Stat("/etc/os-release"); err == nil {
		content, err := os.ReadFile("/etc/os-release")
		if err != nil {
			return "", "", err
		}

		lines := strings.Split(string(content), "\n")
		var id, versionID, prettyName string

		for _, line := range lines {
			if strings.HasPrefix(line, "ID=") {
				id = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				versionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			}
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
		}

		if prettyName != "" {
			return id, prettyName, nil
		}
		return id, versionID, nil
	}
	return "unknown", "unknown", nil
}

func parseBytes(s string) (uint64, error) {
	var multiplier uint64 = 1
	s = strings.ToUpper(s)

	if strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "K")
	}

	var value uint64
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
