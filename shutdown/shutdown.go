// Package shutdown provides graceful shutdown management for Go services.
// It allows registering named closers with priorities and timeouts, and then
// waiting for a termination signal (SIGINT/SIGTERM) to execute them in order.
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//	    "time"
//
//	    "github.com/system-highload-architect/go-solutions/shutdown"
//	)
//
//	func main() {
//	    srv := &http.Server{Addr: ":8080"}
//	    db, _ := sql.Open("postgres", "...")
//
//	    mgr := shutdown.NewManager(30 * time.Second)
//	    mgr.Add("http_server", 0, func(ctx context.Context) error {
//	        return srv.Shutdown(ctx)
//	    }, 5*time.Second)
//	    mgr.Add("database", 1, func(ctx context.Context) error {
//	        return db.Close()
//	    }, 1*time.Second)
//
//	    go func() {
//	        log.Println("server started")
//	        srv.ListenAndServe()
//	    }()
//
//	    mgr.Wait()
//	}
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

// Logger is a minimal interface compatible with *slog.Logger and log.Logger.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// defaultLogger is used when no logger is set.
type defaultLogger struct{}

func (d *defaultLogger) Info(msg string, args ...any)  {}
func (d *defaultLogger) Error(msg string, args ...any) {}

// Closer is a function that performs cleanup and can be cancelled via context.
type Closer func(ctx context.Context) error

// Manager orchestrates graceful shutdown.
type Manager struct {
	mu           sync.Mutex
	closers      []closerItem
	totalTimeout time.Duration
	logger       Logger
}

type closerItem struct {
	name     string
	priority int
	fn       Closer
	timeout  time.Duration
}

// NewManager creates a new Manager with a total timeout for the entire
// shutdown process.
func NewManager(totalTimeout time.Duration) *Manager {
	return &Manager{
		totalTimeout: totalTimeout,
		logger:       &defaultLogger{},
	}
}

// SetLogger assigns a logger to receive shutdown progress messages.
func (m *Manager) SetLogger(l Logger) {
	if l != nil {
		m.logger = l
	}
}

// Add registers a closer with a given priority (lower values run first)
// and a per‑closer timeout.
func (m *Manager) Add(name string, priority int, fn Closer, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append(m.closers, closerItem{
		name:     name,
		priority: priority,
		fn:       fn,
		timeout:  timeout,
	})
}

// Wait blocks until SIGINT or SIGTERM is received, then calls all registered
// closers in priority order. It returns any error that occurred during
// shutdown.
func (m *Manager) Wait() error {
	// 1. Wait for signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	m.logger.Info("shutdown: received termination signal")

	// 2. Sort closers by priority
	m.mu.Lock()
	sort.Slice(m.closers, func(i, j int) bool {
		return m.closers[i].priority < m.closers[j].priority
	})
	closers := m.closers
	m.mu.Unlock()

	// 3. Execute closers with a global deadline
	ctx, cancel := context.WithTimeout(context.Background(), m.totalTimeout)
	defer cancel()

	var lastErr error
	for _, c := range closers {
		m.logger.Info("shutdown: stopping", "name", c.name)
		cCtx, cCancel := context.WithTimeout(ctx, c.timeout)
		err := c.fn(cCtx)
		cCancel()
		if err != nil {
			m.logger.Error("shutdown: error", "name", c.name, "error", err)
			lastErr = err
		}
	}

	m.logger.Info("shutdown: complete")
	return lastErr
}

// Stop немедленно запускает все зарегистрированные closer'ы с учётом
// приоритетов. Сигнал ОС не требуется. Возвращает ошибку, если любой
// closer завершился неудачно.
func (m *Manager) Stop() error {
	m.mu.Lock()
	sort.Slice(m.closers, func(i, j int) bool {
		return m.closers[i].priority < m.closers[j].priority
	})
	closers := m.closers
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.totalTimeout)
	defer cancel()

	var lastErr error
	for _, c := range closers {
		m.logger.Info("shutdown: stopping", "name", c.name)
		cCtx, cCancel := context.WithTimeout(ctx, c.timeout)
		err := c.fn(cCtx)
		cCancel()
		if err != nil {
			m.logger.Error("shutdown: error", "name", c.name, "error", err)
			lastErr = err
		}
	}
	m.logger.Info("shutdown: complete (manual stop)")
	return lastErr
}
