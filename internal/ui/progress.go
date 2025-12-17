package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

// ProgressManager управляет прогресс-индикаторами
type ProgressManager struct{}

// NewSpinner создает новый спиннер
func (pm *ProgressManager) NewSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	return s
}

// NewProgressBar создает новый прогресс-бар
func (pm *ProgressManager) NewProgressBar(total int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
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
}

// ShowProgressWithSpinner показывает прогресс со спиннером
func (pm *ProgressManager) ShowProgressWithSpinner(task func() error, message string) error {
	s := pm.NewSpinner(message)
	s.Start()

	err := task()

	s.Stop()
	if err != nil {
		fmt.Printf("✗ %s: %v\n", message, err)
		return err
	}

	fmt.Printf("✓ %s\n", message)
	return nil
}

// ShowProgressWithBar показывает прогресс с прогресс-баром
func (pm *ProgressManager) ShowProgressWithBar(items []string, processItem func(string) error, description string) error {
	bar := pm.NewProgressBar(len(items), description)

	for _, item := range items {
		if err := processItem(item); err != nil {
			return err
		}
		bar.Add(1)
	}

	bar.Finish()
	return nil
}

// MultiProgress управляет несколькими прогресс-индикаторами
type MultiProgress struct {
	spinners []*spinner.Spinner
	bars     []*progressbar.ProgressBar
}

// NewMultiProgress создает новый MultiProgress
func NewMultiProgress() *MultiProgress {
	return &MultiProgress{
		spinners: make([]*spinner.Spinner, 0),
		bars:     make([]*progressbar.ProgressBar, 0),
	}
}

// AddSpinner добавляет спиннер
func (mp *MultiProgress) AddSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	mp.spinners = append(mp.spinners, s)
	return s
}

// AddProgressBar добавляет прогресс-бар
func (mp *MultiProgress) AddProgressBar(total int, description string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[cyan]=[reset]",
			SaucerHead:    "[cyan]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	mp.bars = append(mp.bars, bar)
	return bar
}

// StartAll запускает все спиннеры
func (mp *MultiProgress) StartAll() {
	for _, s := range mp.spinners {
		s.Start()
	}
}

// StopAll останавливает все спиннеры
func (mp *MultiProgress) StopAll() {
	for _, s := range mp.spinners {
		s.Stop()
	}
}

// UpdateBar обновляет конкретный прогресс-бар
func (mp *MultiProgress) UpdateBar(index int, value int) error {
	if index < 0 || index >= len(mp.bars) {
		return fmt.Errorf("неверный индекс прогресс-бара: %d", index)
	}
	mp.bars[index].Add(value)
	return nil
}

// FinishAll завершает все прогресс-бары
func (mp *MultiProgress) FinishAll() {
	for _, bar := range mp.bars {
		bar.Finish()
	}
}

// ColorProgressBar создает цветной прогресс-бар
func (pm *ProgressManager) ColorProgressBar(total int, description, color string) *progressbar.ProgressBar {
	var saucerColor, headColor string

	switch color {
	case "red":
		saucerColor = "[red]=[reset]"
		headColor = "[red]>[reset]"
	case "green":
		saucerColor = "[green]=[reset]"
		headColor = "[green]>[reset]"
	case "yellow":
		saucerColor = "[yellow]=[reset]"
		headColor = "[yellow]>[reset]"
	case "blue":
		saucerColor = "[blue]=[reset]"
		headColor = "[blue]>[reset]"
	case "cyan":
		saucerColor = "[cyan]=[reset]"
		headColor = "[cyan]>[reset]"
	case "magenta":
		saucerColor = "[magenta]=[reset]"
		headColor = "[magenta]>[reset]"
	default:
		saucerColor = "[green]=[reset]"
		headColor = "[green]>[reset]"
	}

	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        saucerColor,
			SaucerHead:    headColor,
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}

// StepProgress представляет пошаговый прогресс
type StepProgress struct {
	totalSteps int
	currentStep int
	bar        *progressbar.ProgressBar
}

// NewStepProgress создает новый пошаговый прогресс
func (pm *ProgressManager) NewStepProgress(totalSteps int, description string) *StepProgress {
	bar := pm.NewProgressBar(totalSteps, description)
	return &StepProgress{
		totalSteps:  totalSteps,
		currentStep: 0,
		bar:         bar,
	}
}

// NextStep переходит к следующему шагу
func (sp *StepProgress) NextStep() error {
	if sp.currentStep >= sp.totalSteps {
		return fmt.Errorf("достигнут предел шагов")
	}
	sp.currentStep++
	return sp.bar.Add(1)
}

// SetStep устанавливает конкретный шаг
func (sp *StepProgress) SetStep(step int) error {
	if step < 0 || step > sp.totalSteps {
		return fmt.Errorf("неверный номер шага: %d", step)
	}
	
	diff := step - sp.currentStep
	if diff > 0 {
		sp.currentStep = step
		return sp.bar.Add(diff)
	}
	return nil
}

// Finish завершает прогресс
func (sp *StepProgress) Finish() {
	if sp.currentStep < sp.totalSteps {
		sp.bar.Add(sp.totalSteps - sp.currentStep)
	}
	sp.bar.Finish()
}

// AnimatedMessage показывает анимированное сообщение
func (pm *ProgressManager) AnimatedMessage(messages []string, delay time.Duration) {
	for _, msg := range messages {
		s := pm.NewSpinner(msg)
		s.Start()
		time.Sleep(delay)
		s.Stop()
		fmt.Printf("✓ %s\n", msg)
	}
}

// ProgressLogger комбинирует логирование и прогресс
type ProgressLogger struct {
	progressMgr *ProgressManager
}

// NewProgressLogger создает новый ProgressLogger
func NewProgressLogger() *ProgressLogger {
	return &ProgressLogger{
		progressMgr: &ProgressManager{},
	}
}

// LogWithProgress логирует с прогрессом
func (pl *ProgressLogger) LogWithProgress(message string, task func() error) error {
	fmt.Printf("→ %s\n", message)
	return pl.progressMgr.ShowProgressWithSpinner(task, "Выполнение")
}

// ParallelProgress выполняет задачи параллельно с прогрессом
func (pm *ProgressManager) ParallelProgress(tasks []func() error, description string) error {
	type taskResult struct {
		index int
		err   error
	}

	results := make(chan taskResult, len(tasks))
	bar := pm.NewProgressBar(len(tasks), description)

	// Запускаем задачи
	for i, task := range tasks {
		go func(idx int, t func() error) {
			err := t()
			results <- taskResult{index: idx, err: err}
		}(i, task)
	}

	// Собираем результаты
	var errors []error
	for i := 0; i < len(tasks); i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, fmt.Errorf("задача %d: %v", result.index, result.err))
		}
		bar.Add(1)
	}

	bar.Finish()

	if len(errors) > 0 {
		return fmt.Errorf("ошибки в %d задачах: %v", len(errors), errors)
	}

	return nil
}