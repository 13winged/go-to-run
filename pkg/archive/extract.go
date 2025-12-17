package archive

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/briandowns/spinner"
)

// ExtractManager управляет извлечением архивов
type ExtractManager struct{}

// ArchiveInfo содержит информацию об архиве
type ArchiveInfo struct {
	Path     string
	Size     int64
	Type     string
	IsValid  bool
	Contents []string
}

// SupportedFormats возвращает поддерживаемые форматы архивов
func (em *ExtractManager) SupportedFormats() []string {
	return []string{
		".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz",
		".tar", ".gz", ".bz2", ".xz", ".zip", ".rar", ".7z",
		".lz4", ".zst", ".lzop", ".tar.zst", ".tar.lz4",
	}
}

// GetArchiveInfo возвращает информацию об архиве
func (em *ExtractManager) GetArchiveInfo(filepath string) (*ArchiveInfo, error) {
	info := &ArchiveInfo{
		Path: filepath,
	}

	// Получаем размер файла
	if stat, err := os.Stat(filepath); err == nil {
		info.Size = stat.Size()
	}

	// Определяем тип архива
	info.Type = em.detectArchiveType(filepath)

	// Проверяем валидность архива
	info.IsValid = em.checkArchiveValidity(filepath)

	// Получаем список содержимого (если возможно)
	if info.IsValid {
		info.Contents = em.listArchiveContents(filepath)
	}

	return info, nil
}

// Extract извлекает архив
func (em *ExtractManager) Extract(archivePath, outputDir string, showProgress bool) error {
	if !em.isArchive(archivePath) {
		return fmt.Errorf("неподдерживаемый формат архива: %s", archivePath)
	}

	// Создаем директорию для извлечения если не существует
	if outputDir == "" {
		outputDir = em.getDefaultOutputDir(archivePath)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %v", err)
	}

	if showProgress {
		return em.extractWithProgress(archivePath, outputDir)
	}
	return em.extractWithoutProgress(archivePath, outputDir)
}

// ExtractAll извлекает несколько архивов
func (em *ExtractManager) ExtractAll(archives []string, outputDir string, showProgress bool) error {
	if showProgress {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Извлечение %d архивов...", len(archives))
		s.Start()
		defer s.Stop()
	}

	for i, archive := range archives {
		if showProgress {
			fmt.Printf("Извлечение %d/%d: %s\n", i+1, len(archives), filepath.Base(archive))
		}

		subDir := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(archive), filepath.Ext(archive)))
		if err := em.Extract(archive, subDir, false); err != nil {
			return fmt.Errorf("ошибка извлечения %s: %v", archive, err)
		}
	}

	return nil
}

// CreateArchive создает архив
func (em *ExtractManager) CreateArchive(files []string, outputPath string, format string) error {
	switch format {
	case "tar.gz":
		return em.createTarGz(files, outputPath)
	case "zip":
		return em.createZip(files, outputPath)
	case "tar.bz2":
		return em.createTarBz2(files, outputPath)
	case "tar.xz":
		return em.createTarXz(files, outputPath)
	case "7z":
		return em.create7z(files, outputPath)
	default:
		return fmt.Errorf("неподдерживаемый формат: %s", format)
	}
}

// CheckTools проверяет наличие необходимых инструментов
func (em *ExtractManager) CheckTools() map[string]bool {
	tools := map[string]string{
		"tar":    "tar",
		"gzip":   "gzip",
		"bzip2":  "bzip2",
		"xz":     "xz",
		"unzip":  "unzip",
		"unrar":  "unrar",
		"7z":     "7z",
		"lz4":    "lz4",
		"zstd":   "zstd",
		"lzop":   "lzop",
		"gunzip": "gunzip",
	}

	result := make(map[string]bool)
	for name, cmd := range tools {
		result[name] = em.commandExists(cmd)
	}

	return result
}

// Helper методы

func (em *ExtractManager) isArchive(filepath string) bool {
	ext := strings.ToLower(filepath.Ext(filepath))
	for _, format := range em.SupportedFormats() {
		if strings.HasSuffix(filepath, format) {
			return true
		}
	}
	return false
}

func (em *ExtractManager) detectArchiveType(filepath string) string {
	filename := strings.ToLower(filepath)

	switch {
	case strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(filename, ".tar.bz2") || strings.HasSuffix(filename, ".tbz2"):
		return "tar.bz2"
	case strings.HasSuffix(filename, ".tar.xz") || strings.HasSuffix(filename, ".txz"):
		return "tar.xz"
	case strings.HasSuffix(filename, ".tar.zst"):
		return "tar.zst"
	case strings.HasSuffix(filename, ".tar.lz4"):
		return "tar.lz4"
	case strings.HasSuffix(filename, ".tar"):
		return "tar"
	case strings.HasSuffix(filename, ".gz"):
		return "gz"
	case strings.HasSuffix(filename, ".bz2"):
		return "bz2"
	case strings.HasSuffix(filename, ".xz"):
		return "xz"
	case strings.HasSuffix(filename, ".zip"):
		return "zip"
	case strings.HasSuffix(filename, ".rar"):
		return "rar"
	case strings.HasSuffix(filename, ".7z"):
		return "7z"
	case strings.HasSuffix(filename, ".lz4"):
		return "lz4"
	case strings.HasSuffix(filename, ".zst"):
		return "zst"
	case strings.HasSuffix(filename, ".lzop"):
		return "lzop"
	default:
		return "unknown"
	}
}

func (em *ExtractManager) checkArchiveValidity(filepath string) bool {
	archiveType := em.detectArchiveType(filepath)

	switch archiveType {
	case "tar.gz", "tgz", "tar.bz2", "tbz2", "tar.xz", "txz", "tar":
		cmd := exec.Command("tar", "-tf", filepath)
		return cmd.Run() == nil
	case "gz":
		cmd := exec.Command("gunzip", "-t", filepath)
		return cmd.Run() == nil
	case "zip":
		cmd := exec.Command("unzip", "-t", filepath)
		return cmd.Run() == nil
	case "rar":
		if em.commandExists("unrar") {
			cmd := exec.Command("unrar", "t", filepath)
			return cmd.Run() == nil
		}
		return true // Предполагаем валидным если нет unrar
	default:
		return true // Для остальных форматов считаем валидным
	}
}

func (em *ExtractManager) listArchiveContents(filepath string) []string {
	archiveType := em.detectArchiveType(filepath)

	switch archiveType {
	case "tar.gz", "tgz", "tar.bz2", "tbz2", "tar.xz", "txz", "tar":
		cmd := exec.Command("tar", "-tf", filepath)
		if output, err := cmd.Output(); err == nil {
			return strings.Split(strings.TrimSpace(string(output)), "\n")
		}
	case "zip":
		cmd := exec.Command("unzip", "-l", filepath)
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 3 {
				return lines[3 : len(lines)-3]
			}
		}
	}

	return []string{}
}

func (em *ExtractManager) getDefaultOutputDir(archivePath string) string {
	filename := filepath.Base(archivePath)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	return filepath.Join(filepath.Dir(archivePath), baseName)
}

func (em *ExtractManager) extractWithProgress(archivePath, outputDir string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Извлечение архива..."
	s.Start()
	defer s.Stop()

	return em.extractArchive(archivePath, outputDir)
}

func (em *ExtractManager) extractWithoutProgress(archivePath, outputDir string) error {
	return em.extractArchive(archivePath, outputDir)
}

func (em *ExtractManager) extractArchive(archivePath, outputDir string) error {
	archiveType := em.detectArchiveType(archivePath)

	switch archiveType {
	case "tar.gz", "tgz":
		return em.extractTarGz(archivePath, outputDir)
	case "tar.bz2", "tbz2":
		return em.extractTarBz2(archivePath, outputDir)
	case "tar.xz", "txz":
		return em.extractTarXz(archivePath, outputDir)
	case "tar":
		return em.extractTar(archivePath, outputDir)
	case "gz":
		return em.extractGz(archivePath, outputDir)
	case "bz2":
		return em.extractBz2(archivePath, outputDir)
	case "xz":
		return em.extractXz(archivePath, outputDir)
	case "zip":
		return em.extractZip(archivePath, outputDir)
	case "rar":
		return em.extractRar(archivePath, outputDir)
	case "7z":
		return em.extract7z(archivePath, outputDir)
	case "lz4":
		return em.extractLz4(archivePath, outputDir)
	case "zst":
		return em.extractZstd(archivePath, outputDir)
	case "lzop":
		return em.extractLzop(archivePath, outputDir)
	case "tar.zst":
		return em.extractTarZstd(archivePath, outputDir)
	case "tar.lz4":
		return em.extractTarLz4(archivePath, outputDir)
	default:
		return fmt.Errorf("неподдерживаемый формат архива: %s", archiveType)
	}
}

// Методы извлечения для разных форматов

func (em *ExtractManager) extractTarGz(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractTarBz2(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "-xjf", archivePath, "-C", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractTarXz(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "-xJf", archivePath, "-C", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractTar(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "-xf", archivePath, "-C", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractGz(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".gz"))
	
	cmd := exec.Command("gunzip", "-c", archivePath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputFile, output, 0644)
}

func (em *ExtractManager) extractBz2(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".bz2"))
	
	cmd := exec.Command("bunzip2", "-c", archivePath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputFile, output, 0644)
}

func (em *ExtractManager) extractXz(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".xz"))
	
	cmd := exec.Command("xz", "-d", "-c", archivePath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	return os.WriteFile(outputFile, output, 0644)
}

func (em *ExtractManager) extractZip(archivePath, outputDir string) error {
	cmd := exec.Command("unzip", "-o", archivePath, "-d", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractRar(archivePath, outputDir string) error {
	cmd := exec.Command("unrar", "x", archivePath, outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extract7z(archivePath, outputDir string) error {
	cmd := exec.Command("7z", "x", archivePath, "-o"+outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractLz4(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".lz4"))
	
	cmd := exec.Command("lz4", "-d", archivePath, outputFile)
	return cmd.Run()
}

func (em *ExtractManager) extractZstd(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".zst"))
	
	cmd := exec.Command("zstd", "-d", archivePath, "-o", outputFile)
	return cmd.Run()
}

func (em *ExtractManager) extractLzop(archivePath, outputDir string) error {
	filename := filepath.Base(archivePath)
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(filename, ".lzop"))
	
	cmd := exec.Command("lzop", "-d", archivePath, "-o", outputFile)
	return cmd.Run()
}

func (em *ExtractManager) extractTarZstd(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "--zstd", "-xf", archivePath, "-C", outputDir)
	return cmd.Run()
}

func (em *ExtractManager) extractTarLz4(archivePath, outputDir string) error {
	cmd := exec.Command("tar", "--lz4", "-xf", archivePath, "-C", outputDir)
	return cmd.Run()
}

// Методы создания архивов

func (em *ExtractManager) createTarGz(files []string, outputPath string) error {
	args := []string{"-czf", outputPath}
	args = append(args, files...)
	cmd := exec.Command("tar", args...)
	return cmd.Run()
}

func (em *ExtractManager) createZip(files []string, outputPath string) error {
	args := []string{outputPath}
	args = append(args, files...)
	cmd := exec.Command("zip", args...)
	return cmd.Run()
}

func (em *ExtractManager) createTarBz2(files []string, outputPath string) error {
	args := []string{"-cjf", outputPath}
	args = append(args, files...)
	cmd := exec.Command("tar", args...)
	return cmd.Run()
}

func (em *ExtractManager) createTarXz(files []string, outputPath string) error {
	args := []string{"-cJf", outputPath}
	args = append(args, files...)
	cmd := exec.Command("tar", args...)
	return cmd.Run()
}

func (em *ExtractManager) create7z(files []string, outputPath string) error {
	args := []string{"a", outputPath}
	args = append(args, files...)
	cmd := exec.Command("7z", args...)
	return cmd.Run()
}

func (em *ExtractManager) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ExtractFunction предоставляет функцию извлечения для использования в скриптах
func ExtractFunction() func(string) error {
	return func(archivePath string) error {
		em := &ExtractManager{}
		return em.Extract(archivePath, "", true)
	}
}

// ServiceInfo представляет информацию о службе (для table.go)
type ServiceInfo struct {
	Name        string
	Status      string
	AutoStart   bool
	Description string
}

// sortStrings сортирует строки (для table.go)
func sortStrings(strings []string) {
	sort.Strings(strings)
}