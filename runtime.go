package main

import (
	"context"
	"os/exec"

	"go.uber.org/zap"
)

type Runtime interface {
	// Returns the the compiler's exit code, program's exit code, and an error
	// if one occurred.
	CompileAndRun(ctx context.Context, logger *zap.Logger, code string) (int, int, error)
}

type PythonRuntime struct {
	logger *zap.Logger
}

var _ Runtime = &PythonRuntime{}

func NewPythonRuntime(logger *zap.Logger) *PythonRuntime {
	return &PythonRuntime{
		logger: logger,
	}
}

func (r *PythonRuntime) CompileAndRun(ctx context.Context, logger *zap.Logger, code string) (int, int, error) {
	cmd := exec.CommandContext(ctx, "python3", "-c", code)

	if err := cmd.Start(); err != nil {
		return -1, -1, err
	}

	logger.Info("running program")

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Exited() {
				logger.Info("program exited with non-zero exit code", zap.Int("code", exitErr.ExitCode()))
				return 0, exitErr.ExitCode(), nil
			}
			if ctx.Err() == context.DeadlineExceeded {
				logger.Info("program timed out", zap.Error(ctx.Err()))
				return -1, -1, ctx.Err()
			}
			logger.Error("program was terminated", zap.Error(err))
			return -1, -1, err
		}
		logger.Error("program failed to run", zap.Error(err))
		return 0, -1, err
	}

	logger.Info("program ran successfully")

	return 0, 0, nil
}
