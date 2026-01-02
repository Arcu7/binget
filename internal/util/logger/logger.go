package logger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/fatih/color"
)

// DevHandler is a custom slog handler for development logging.
// It uses fatih/color to colorize log levels (not hand rolling my own because i have no time).
type DevHandler struct {
	handler slog.Handler // Go official documentation does not recommend this, but oh well
	b       *bytes.Buffer
	mu      *sync.Mutex
}

func (l *DevHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return l.handler.Enabled(ctx, level)
}

func (l *DevHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DevHandler{handler: l.handler.WithAttrs(attrs), b: l.b, mu: l.mu}
}

func (l *DevHandler) WithGroup(name string) slog.Handler {
	return &DevHandler{handler: l.handler.WithGroup(name), b: l.b, mu: l.mu}
}

func (l *DevHandler) Handle(ctx context.Context, record slog.Record) error {
	level := record.Level.String() + ":"

	switch record.Level {
	case slog.LevelDebug:
		level = color.MagentaString("DEBUG:")
	case slog.LevelInfo:
		level = color.GreenString("INFO:")
	case slog.LevelWarn:
		level = color.YellowString("WARN:")
	case slog.LevelError:
		level = color.RedString("ERROR:")
	}

	attrs, err := l.computeAttrs(ctx, record)
	if err != nil {
		return fmt.Errorf("error when computing attrs: %w", err)
	}

	fmt.Printf("%s %s %s", level, record.Message, attrs)

	return nil
}

func (l *DevHandler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (string, error) {
	l.mu.Lock()
	defer func() {
		l.b.Reset()
		l.mu.Unlock()
	}()
	if err := l.handler.Handle(ctx, r); err != nil {
		return "", fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	attrs := l.b.String()

	return attrs, nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func NewDevLogger(opts *slog.HandlerOptions) *DevHandler {
	b := &bytes.Buffer{}
	return &DevHandler{
		b: b,
		handler: slog.NewTextHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		mu: &sync.Mutex{},
	}
}

func New(verbose bool) *slog.Logger {
	var handler slog.Handler
	if !verbose {
		handler = slog.DiscardHandler
	} else {
		handler = NewDevLogger(&slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
