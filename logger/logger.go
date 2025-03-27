package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	// Colors
	successColor   = lipgloss.Color("#04B575") // Green
	errorColor     = lipgloss.Color("#FF4365") // Red
	warnColor      = lipgloss.Color("#FF8E3C") // Orange
	infoColor      = lipgloss.Color("#2D7DD2") // Blue
	highlightColor = lipgloss.Color("#9e8ad7") // Purple

	// Styles
	successStyle   = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	warnStyle      = lipgloss.NewStyle().Foreground(warnColor).Bold(true)
	infoStyle      = lipgloss.NewStyle().Foreground(infoColor).Bold(true)
	highlightStyle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true)

	// Prefixes
	successPrefix = successStyle.Render("SUCCESS")
	errorPrefix   = errorStyle.Render("ERROR")
	warnPrefix    = warnStyle.Render("WARNING")
	infoPrefix    = infoStyle.Render("INFO")
	taskPrefix    = highlightStyle.Render("TASK")
	rulePrefix    = highlightStyle.Render("RULE")
	dbPrefix      = highlightStyle.Render("DB")
)

// Logger is a wrapper around charm log
type Logger struct {
	logger *log.Logger
}

func New() *Logger {
	l := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04",
	})

	return &Logger{
		logger: l,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Success(msg string, args ...interface{}) {
	l.logger.SetPrefix(successPrefix)
	l.logger.Info(msg, args...)
	l.logger.SetPrefix("")
}

func (l *Logger) Error(err error, msg string, args ...interface{}) {
	l.logger.SetPrefix(errorPrefix)
	l.logger.Error(msg, append(args, "error", err)...)
	l.logger.SetPrefix("")
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.SetPrefix(warnPrefix)
	l.logger.Warn(msg, args...)
	l.logger.SetPrefix("")
}

func (l *Logger) Task(name string, msg string) {
	l.logger.SetPrefix(taskPrefix)
	l.logger.Info(fmt.Sprintf("%s: %s", highlightStyle.Render(name), msg))
	l.logger.SetPrefix("")
}

func (l *Logger) Rule(name string, msg string) {
	l.logger.SetPrefix(rulePrefix)
	l.logger.Info(fmt.Sprintf("%s: %s", highlightStyle.Render(name), msg))
	l.logger.SetPrefix("")
}

func (l *Logger) DB(name string, msg string) {
	l.logger.SetPrefix(dbPrefix)
	l.logger.Info(fmt.Sprintf("%s: %s", highlightStyle.Render(name), msg))
	l.logger.SetPrefix("")
}

// Result logs query results with nice formatting
func (l *Logger) Result(ruleName string, result string) {
	l.logger.SetPrefix(successPrefix)
	l.logger.Info(fmt.Sprintf("Results for %s:", highlightStyle.Render(ruleName)))
	fmt.Println(result)
	l.logger.SetPrefix("")
}

var DefaultLogger = New()

// Expose global logger functions
func Info(msg string, args ...interface{})             { DefaultLogger.Info(msg, args...) }
func Success(msg string, args ...interface{})          { DefaultLogger.Success(msg, args...) }
func Error(err error, msg string, args ...interface{}) { DefaultLogger.Error(err, msg, args...) }
func Warn(msg string, args ...interface{})             { DefaultLogger.Warn(msg, args...) }
func Task(name string, msg string)                     { DefaultLogger.Task(name, msg) }
func Rule(name string, msg string)                     { DefaultLogger.Rule(name, msg) }
func DB(name string, msg string)                       { DefaultLogger.DB(name, msg) }
func Result(ruleName string, result string)            { DefaultLogger.Result(ruleName, result) }
