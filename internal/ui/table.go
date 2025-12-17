package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// TableManager управляет таблицами
type TableManager struct{}

// NewTable создает новую таблицу
func (tm *TableManager) NewTable(headers []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	return table
}

// NewBorderedTable создает таблицу с рамкой
func (tm *TableManager) NewBorderedTable(headers []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(true)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)
	return table
}

// NewColorTable создает цветную таблицу
func (tm *TableManager) NewColorTable(headers []string, headerColors []tablewriter.Colors, columnColors []tablewriter.Colors) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	// Устанавливаем цвета заголовков
	if len(headerColors) > 0 {
		table.SetHeaderColor(headerColors...)
	}

	// Устанавливаем цвета колонок
	if len(columnColors) > 0 {
		table.SetColumnColor(columnColors...)
	}

	return table
}

// DisplaySystemInfo отображает информацию о системе в таблице
func (tm *TableManager) DisplaySystemInfo(info map[string]string) {
	table := tm.NewTable([]string{"Параметр", "Значение"})
	table.SetColumnSeparator(":")
	table.SetAutoWrapText(false)

	// Сортируем ключи для красивого вывода
	var keys []string
	for k := range info {
		keys = append(keys, k)
	}
	sortStrings(keys)

	// Добавляем данные
	for _, key := range keys {
		value := info[key]
		// Обрезаем длинные значения
		if len(value) > 80 {
			value = value[:77] + "..."
		}
		table.Append([]string{key, value})
	}

	table.Render()
}

// DisplayPackages отображает список пакетов в таблице
func (tm *TableManager) DisplayPackages(packages []string, category string) {
	if len(packages) == 0 {
		fmt.Printf("Нет пакетов в категории: %s\n", category)
		return
	}

	table := tm.NewBorderedTable([]string{"#", "Пакет", "Категория"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)

	for i, pkg := range packages {
		table.Append([]string{fmt.Sprintf("%d", i+1), pkg, category})
	}

	fmt.Printf("\nПакеты в категории '%s':\n", category)
	table.Render()
}

// DisplayCategories отображает категории пакетов
func (tm *TableManager) DisplayCategories(categories map[string][]string) {
	table := tm.NewColorTable(
		[]string{"Категория", "Кол-во пакетов", "Описание"},
		[]tablewriter.Colors{
			{tablewriter.Bold, tablewriter.BgBlueColor},
			{tablewriter.Bold, tablewriter.BgGreenColor},
			{tablewriter.Bold, tablewriter.BgCyanColor},
		},
		[]tablewriter.Colors{
			{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, tablewriter.FgHiGreenColor},
			{tablewriter.FgHiCyanColor},
		},
	)

	descriptions := map[string]string{
		"basic":       "Основные утилиты системы",
		"archive":     "Инструменты для работы с архивами",
		"network":     "Сетевые утилиты и инструменты",
		"monitoring":  "Мониторинг системы",
		"development": "Инструменты разработки",
		"security":    "Безопасность системы",
		"system":      "Системные утилиты",
		"database":    "Базы данных",
		"web":         "Веб-серверы и инструменты",
	}

	for category, packages := range categories {
		desc := descriptions[category]
		if desc == "" {
			desc = "Без описания"
		}
		table.Append([]string{
			strings.Title(category),
			fmt.Sprintf("%d", len(packages)),
			desc,
		})
	}

	fmt.Println("\nДоступные категории пакетов:")
	table.Render()
}

// DisplayServices отображает список служб
func (tm *TableManager) DisplayServices(services []ServiceInfo) {
	table := tm.NewBorderedTable([]string{"Служба", "Статус", "Автозагрузка", "Описание"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgCyanColor},
	)

	for _, service := range services {
		statusColor := tablewriter.FgHiRedColor
		if service.Status == "active" {
			statusColor = tablewriter.FgHiGreenColor
		} else if service.Status == "inactive" {
			statusColor = tablewriter.FgHiYellowColor
		}

		autoStart := "❌"
		if service.AutoStart {
			autoStart = "✅"
		}

		table.Rich([]string{
			service.Name,
			service.Status,
			autoStart,
			service.Description,
		}, []tablewriter.Colors{
			{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, statusColor},
			{},
			