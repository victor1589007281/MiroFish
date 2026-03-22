package simulation

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"swarm-predict/internal/model"
)

// ActionLog records agent actions per round.
type ActionLog struct {
	Round     int            `json:"round"`
	Timestamp time.Time      `json:"timestamp"`
	Decision  model.Decision `json:"decision"`
}

// Logger writes simulation actions to a JSONL file.
type Logger struct {
	mu   sync.Mutex
	file *os.File
	logs []ActionLog
}

func NewLogger(path string) (*Logger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}
	return &Logger{file: f}, nil
}

// NewInMemoryLogger creates a logger that only stores in memory (for testing).
func NewInMemoryLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Log(round int, d model.Decision) {
	entry := ActionLog{
		Round:     round,
		Timestamp: time.Now(),
		Decision:  d,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, entry)
	if l.file != nil {
		data, _ := json.Marshal(entry)
		l.file.Write(data)
		l.file.Write([]byte("\n"))
	}
}

func (l *Logger) GetLogs() []ActionLog {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]ActionLog{}, l.logs...)
}

// GetLogsByRound returns action logs filtered to a specific round.
func (l *Logger) GetLogsByRound(round int) []ActionLog {
	l.mu.Lock()
	defer l.mu.Unlock()
	var result []ActionLog
	for _, log := range l.logs {
		if log.Round == round {
			result = append(result, log)
		}
	}
	return result
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
