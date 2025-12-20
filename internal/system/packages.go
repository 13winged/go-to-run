// Package system предоставляет функциональность для работы с системой.
package system

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

// PackageManager представляет менеджер пакетов
type PackageManager struct {
	Name    string
	Update  string
	Upgrade string
	Install string
	Remove  string
	Clean   string
	Check   string
}

// PackageCategory представляет категорию пакетов
type PackageCategory struct {
	Name     string
	Packages []string
	Enabled  bool
}

// PackageManagerDetector определяет менеджер пакетов
type PackageManagerDetector struct{}

var (
	packageManagers = map[string]PackageManager{
		"apt": {
			Name:    "apt",
			Update:  "apt update",
			Upgrade: "apt upgrade -y",
			Install: "apt install -y",
			Remove:  "apt remove -y",
			Clean:   "apt autoremove -y && apt autoclean",
			Check:   "apt list --upgradable",
		},
		"dnf": {
			Name:    "dnf",
			Update:  "dnf check-update",
			Upgrade: "dnf update -y",
			Install: "dnf install -y",
			Remove:  "dnf remove -y",
			Clean:   "dnf clean all",
			Check:   "dnf check-update",
		},
		"yum": {
			Name:    "yum",
			Update:  "yum check-update",
			Upgrade: "yum update -y",
			Install: "yum install -y",
			Remove:  "yum remove -y",
			Clean:   "yum clean all",
			Check:   "yum check-update",
		},
		"pacman": {
			Name:    "pacman",
			Update:  "pacman -Sy",
			Upgrade: "pacman -Syu --noconfirm",
			Install: "pacman -S --noconfirm",
			Remove:  "pacman -R --noconfirm",
			Clean:   "pacman -Sc --noconfirm",
			Check:   "pacman -Qu",
		},
		"apk": {
			Name:    "apk",
			Update:  "apk update",
			Upgrade: "apk upgrade",
			Install: "apk add",
			Remove:  "apk del",
			Clean:   "apk cache clean",
			Check:   "apk version",
		},
		"zypper": {
			Name:    "zypper",
			Update:  "zypper refresh",
			Upgrade: "zypper update -y",
			Install: "zypper install -y",
			Remove:  "zypper remove -y",
			Clean:   "zypper clean",
			Check:   "zypper list-updates",
		},
	}

	packageCategories = map[string]PackageCategory{
		"basic": {
			Name: "Basic Utilities",
			Packages: []string{
				"nano", "vim", "micro",
				"htop", "btop", "glances",
				"git", "curl", "wget", "rsync",
				"tree", "tmux", "screen", "zsh",
				"sudo", "ca-certificates",
			},
			Enabled: true,
		},
		"archive": {
			Name: "Archive Tools",
			Packages: []string{
				"gzip", "gunzip", "zip", "unzip",
				"p7zip-full", "p7zip-rar", "unrar",
				"bzip2", "xz-utils", "zstd",
				"lz4", "tar", "cpio", "lzop",
				"lbzip2", "pigz", "pbzip2",
			},
			Enabled: true,
		},
	}
)

// Detect определяет менеджер пакетов системы
func (d *PackageManagerDetector) Detect() (*PackageManager, error) {
	for cmd, pm := range packageManagers {
		if commandExists(cmd) {
			return &pm, nil
		}
	}
	return nil, errors.New("не найден поддерживаемый менеджер пакетов")
}

// IsPackageInstalled проверяет установлен ли пакет
func IsPackageInstalled(pm *PackageManager, pkg string) (bool, error) {
	switch pm.Name {
	case "apt":
		cmd := "dpkg-query -W -f='${Status}' " + pkg + " 2>/dev/null | grep -q 'install ok installed'"
		_, err := exec.Command("sh", "-c", cmd).Output()
		return err == nil, nil
	case "dnf", "yum":
		cmd := "rpm -q " + pkg
		_, err := exec.Command("sh", "-c", cmd).Output()
		return err == nil, nil
	case "pacman":
		cmd := "pacman -Qs ^" + pkg + "$"
		output, err := exec.Command("sh", "-c", cmd).Output()
		return err == nil && strings.Contains(string(output), pkg), nil
	case "apk":
		cmd := "apk info -e " + pkg
		_, err := exec.Command("sh", "-c", cmd).Output()
		return err == nil, nil
	case "zypper":
		cmd := "rpm -q " + pkg
		_, err := exec.Command("sh", "-c", cmd).Output()
		return err == nil, nil
	default:
		return false, fmt.Errorf("неподдерживаемый менеджер пакетов: %s", pm.Name)
	}
}

// InstallPackages устанавливает пакеты
func InstallPackages(pm *PackageManager, packages []string, showProgress bool) error {
	if len(packages) == 0 {
		return nil
	}

	// Фильтруем уже установленные пакеты
	var toInstall []string
	for _, pkg := range packages {
		installed, err := IsPackageInstalled(pm, pkg)
		if err != nil {
			return fmt.Errorf("ошибка проверки пакета %s: %w", pkg, err)
		}
		if !installed {
			toInstall = append(toInstall, pkg)
		}
	}

	if len(toInstall) == 0 {
		return nil
	}

	if showProgress {
		return installWithProgress(pm, toInstall)
	}
	return installWithoutProgress(pm, toInstall)
}

func installWithProgress(pm *PackageManager, packages []string) error {
	bar := progressbar.NewOptions(len(packages),
		progressbar.OptionSetDescription("Установка пакетов"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	// Для некоторых менеджеров устанавливаем все сразу
	if pm.Name == "apt" || pm.Name == "dnf" || pm.Name == "yum" {
		installCmd := pm.Install + " " + strings.Join(packages, " ")
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Установка пакетов..."
		s.Start()

		cmd := exec.Command("sh", "-c", installCmd)
		if err := cmd.Run(); err != nil {
			s.Stop()
			// Пробуем установить по одному
			for _, pkg := range packages {
				cmdStr := pm.Install + " " + pkg
				cmd := exec.Command("sh", "-c", cmdStr)
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("ошибка установки %s: %w", pkg, err)
				}
				// Игнорируем ошибки прогресс-бара
				_ = bar.Add(1)
			}
		} else {
			s.Stop()
			// Игнорируем ошибки прогресс-бара
			_ = bar.Add(len(packages))
		}
	} else {
		// Для других менеджеров устанавливаем по одному
		for _, pkg := range packages {
			cmdStr := pm.Install + " " + pkg
			cmd := exec.Command("sh", "-c", cmdStr)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("ошибка установки %s: %w", pkg, err)
			}
			// Игнорируем ошибки прогресс-бара
			_ = bar.Add(1)
		}
	}

	// Игнорируем ошибки завершения
	_ = bar.Finish()
	return nil
}

func installWithoutProgress(pm *PackageManager, packages []string) error {
	if pm.Name == "apt" || pm.Name == "dnf" || pm.Name == "yum" {
		cmdStr := pm.Install + " " + strings.Join(packages, " ")
		cmd := exec.Command("sh", "-c", cmdStr)
		return cmd.Run()
	}

	// Для других менеджеров устанавливаем по одному
	for _, pkg := range packages {
		cmdStr := pm.Install + " " + pkg
		cmd := exec.Command("sh", "-c", cmdStr)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ошибка установки %s: %w", pkg, err)
		}
	}
	return nil
}

// UpdateSystem обновляет систему
func UpdateSystem(pm *PackageManager) error {
	// Обновляем список пакетов
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Обновление списка пакетов..."
	s.Start()

	updateCmd := exec.Command("sh", "-c", pm.Update)
	if err := updateCmd.Run(); err != nil {
		s.Stop()
		return fmt.Errorf("ошибка обновления списка пакетов: %w", err)
	}
	s.Stop()

	// Обновляем пакеты
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Обновление пакетов..."
	s.Start()

	upgradeCmd := exec.Command("sh", "-c", pm.Upgrade)
	if err := upgradeCmd.Run(); err != nil {
		s.Stop()
		return fmt.Errorf("ошибка обновления пакетов: %w", err)
	}
	s.Stop()

	return nil
}

// CleanSystem очищает систему
func CleanSystem(pm *PackageManager) error {
	cleanCmd := exec.Command("sh", "-c", pm.Clean)
	return cleanCmd.Run()
}

// GetAvailableUpdates возвращает список доступных обновлений
func GetAvailableUpdates(pm *PackageManager) ([]string, error) {
	checkCmd := exec.Command("sh", "-c", pm.Check)
	output, err := checkCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения обновлений: %w", err)
	}

	var updates []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Listing...") {
			updates = append(updates, line)
		}
	}

	return updates, nil
}

// GetPackageCategories возвращает список категорий пакетов
func GetPackageCategories() []PackageCategory {
	var categories []PackageCategory
	for _, category := range packageCategories {
		categories = append(categories, category)
	}

	// Сортируем по имени
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

// GetPackagesByCategory возвращает пакеты по категории
func GetPackagesByCategory(category string) ([]string, error) {
	cat, ok := packageCategories[category]
	if !ok {
		return nil, fmt.Errorf("категория не найдена: %s", category)
	}
	return cat.Packages, nil
}

// FilterInstalledPackages фильтрует установленные пакеты
func FilterInstalledPackages(pm *PackageManager, packages []string) ([]string, []string, error) {
	var installed, notInstalled []string

	for _, pkg := range packages {
		isInstalled, err := IsPackageInstalled(pm, pkg)
		if err != nil {
			return nil, nil, fmt.Errorf("ошибка проверки пакета %s: %w", pkg, err)
		}
		if isInstalled {
			installed = append(installed, pkg)
		} else {
			notInstalled = append(notInstalled, pkg)
		}
	}

	return installed, notInstalled, nil
}

// InstallCategory устанавливает все пакеты из категории
func InstallCategory(pm *PackageManager, category string, showProgress bool) error {
	packages, err := GetPackagesByCategory(category)
	if err != nil {
		return err
	}
	return InstallPackages(pm, packages, showProgress)
}

// commandExists проверяет существование команды
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
