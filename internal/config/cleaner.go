package config

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Callable interface {
	Invoke(ctx context.Context) error
}

type Cleaner struct {
	cleaners       []Callable
	mu             sync.Mutex
	initOnce       sync.Once
	cleaning       bool
	loggerShutdown Callable
}

var cleanerInstance = &Cleaner{}

func GetCleaner() *Cleaner {
	return cleanerInstance
}

func ReverseForEach[T any](slice []T, f func(index int, value T)) {
	for i := len(slice) - 1; i >= 0; i-- {
		f(i, slice[i])
	}
}

func (c *Cleaner) Add(callable Callable) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cleaning {
		Debug("Cleaner is already shutting down, ignoring new cleaner")
		return
	}
	c.cleaners = append(c.cleaners, callable)
	DebugF("Adding cleaner #%d (%T)", len(c.cleaners), callable)
}

func (c *Cleaner) Clean() {
	c.mu.Lock()
	c.cleaning = true // 标记为清理中，阻止后续Add操作
	cleanersCopy := make([]Callable, len(c.cleaners))
	copy(cleanersCopy, c.cleaners)
	c.mu.Unlock()

	DebugF("Starting cleanup of %d registered functions", len(cleanersCopy))

	var errs []error
	ReverseForEach(cleanersCopy, func(idx int, c Callable) { // 使用匿名函数确保defer在每次迭代执行
		DebugF("Invoking cleaner #%d (%T)", idx+1, c)
		timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc() // 确保每次调用后取消上下文
		if err := c.Invoke(timeoutCtx); err != nil {
			ErrorF("Cleaner #%d (%T) failed: %v", idx+1, c, err) // 记录类型和错误
			errs = append(errs, err)
		}
	})

	if len(errs) > 0 {
		ErrorF("%d errors occurred during cleanup:", len(errs))
		for i, err := range errs {
			ErrorF("Error %d: %v", i+1, err)
		}
	} else {
		Debug("All cleaners executed successfully")
	}
	Info("Cleanup finished, server offline")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.loggerShutdown.Invoke(shutdownCtx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "LOGGER SHUTDOWN ERROR: %v\n", err)
	}
	syscall.Exit(0)
}

func (c *Cleaner) Init(loggerShutdown Callable) {
	c.initOnce.Do(func() {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		c.loggerShutdown = loggerShutdown
		go func() {
			<-ctx.Done()
			stop()
			Info("Received interrupt signal, shutting down")

			c.Clean()
		}()
	})
}
