package graceful_shutdown

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const shutdownTimeout = 7 * time.Minute

type closeFunc func(ctx context.Context) error

type GracefulShutdown struct {
	mu         sync.Mutex
	ctx        context.Context
	errGroup   *errgroup.Group
	closeFuncs []closeFunc
}

func NewGracefulShutdown(parentCtx context.Context) *GracefulShutdown {
	g, ctx := errgroup.WithContext(parentCtx)
	gfl := &GracefulShutdown{
		ctx:      ctx,
		errGroup: g,
	}

	gfl.Go(gfl.listenerOS)
	gfl.Go(gfl.killer)

	return gfl
}

func (g *GracefulShutdown) Go(foo func() error) {
	g.errGroup.Go(func() (err error) {
		defer func() {
			if errPanic := recover(); errPanic != nil {
				err = fmt.Errorf("panic in graceful shutdown: %v", errPanic)
				slog.Error("panic in graceful shutdown", "error", err)
			}
		}()

		err = foo()

		return err

	})
}

func (g *GracefulShutdown) Wait() {
	err := g.errGroup.Wait()
	if err != nil {
		slog.Error("error in graceful shutdown", "error", err)
	}
}

func (g *GracefulShutdown) MustClose(f closeFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.closeFuncs = append(g.closeFuncs, f)
}

func (g *GracefulShutdown) listenerOS() error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case <-g.ctx.Done():
		return nil
	case signalFromOS := <-ch:
		slog.Info("received signal from OS", "signal", signalFromOS)
		return errors.New("received signal from OS: " + signalFromOS.String())
	}
}

func (g *GracefulShutdown) killer() error {
	<-g.ctx.Done()

	ctx, cancelTimeout := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelTimeout()

	if err := g.close(ctx); err != nil {
		panic(err)
	}

	return nil
}

func (g *GracefulShutdown) close(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	closeErrMessages := make([]string, 0, len(g.closeFuncs))
	complete := make(chan struct{}, 1)

	go func() {
		for _, closeFunc := range g.closeFuncs {
			if err := closeFunc(g.ctx); err != nil {
				closeErrMessages = append(closeErrMessages, fmt.Sprintf("error closing: %v", err))
			}
		}
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("timeout closing")
	}

	if len(closeErrMessages) > 0 {
		return fmt.Errorf("errors closing: %v", strings.Join(closeErrMessages, ", "))
	}

	return nil
}
