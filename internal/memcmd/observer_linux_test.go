//go:build linux

package memcmd

import (
	"bytes"
	"context"
	"io/fs"
	"runtime"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/conc/pool"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestObserverIntegration(t *testing.T) {
	defer goleak.VerifyNone(t)

	cmd := allocatingGoProgram(t, 250*1024*1024) // 250 MB

	var buf bytes.Buffer
	cmd.Stderr = &buf
	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start test program: %v, stdErr: %s", err, buf.String())
	}

	observer, err := NewLinuxObserver(context.Background(), cmd, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create observer: %v", err)
	}

	observer.Start()
	defer observer.Stop()

	err = cmd.Wait()
	if err != nil {
		t.Fatalf("failed to wait for test program: %v stdErr: %s", err, buf.String())
	}

	memoryUsage, err := observer.MaxMemoryUsage()
	if err != nil {
		t.Fatalf("failed to get memory usage: %v", err)
	}

	t.Logf("memory usage: %s", humanize.Bytes(uint64(memoryUsage)))

	memoryLow := bytesize.Size(200 << 20)  // 200 MB
	memoryHigh := bytesize.Size(350 << 20) // 350 MB

	if !(memoryLow < memoryUsage && memoryUsage < memoryHigh) {
		t.Fatalf("memory usage is not in the expected range (low: %s, high: %s): %s", humanize.Bytes(uint64(memoryLow)), humanize.Bytes(uint64(memoryHigh)), humanize.Bytes(uint64(memoryUsage)))
	}
}
func TestErrorMaybeCausedByExplicitStop(t *testing.T) {
	t.Run("normal context cancellation error", func(t *testing.T) {
		explicitlyStopped := make(chan struct{})

		require.False(t, errMaybeCausedByExplicitStop(context.Canceled, explicitlyStopped))
	})

	t.Run("explicit stop", func(t *testing.T) {
		explicitlyStopped := make(chan struct{})
		close(explicitlyStopped)

		require.True(t, errMaybeCausedByExplicitStop(context.Canceled, explicitlyStopped))
	})

	t.Run("stopped, but not cancellation error", func(t *testing.T) {
		explicitlyStopped := make(chan struct{})
		close(explicitlyStopped)

		require.False(t, errMaybeCausedByExplicitStop(errors.New("some error"), explicitlyStopped))
	})

	t.Run("nil error", func(t *testing.T) {
		explicitlyStopped := make(chan struct{})
		close(explicitlyStopped)

		require.False(t, errMaybeCausedByExplicitStop(nil, explicitlyStopped))
	})
}

func TestConvertESRCH(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "Non-ESRCH syscall error",
			err:      syscall.ENOENT,
			expected: syscall.ENOENT,
		},
		{
			name:     "ESRCH error",
			err:      syscall.ESRCH,
			expected: errors.Append(syscall.ESRCH, fs.ErrNotExist),
		},
		{
			name:     "Wrapped ESRCH error",
			err:      errors.Wrap(syscall.ESRCH, "wrapped error"),
			expected: errors.Append(errors.Wrap(syscall.ESRCH, "wrapped error"), fs.ErrNotExist),
		},
		{
			name: "Path error including ESRCH",
			err: &fs.PathError{
				Op:   "open",
				Path: "/proc/1234",
				Err:  syscall.ESRCH,
			},
			expected: errors.Append(&fs.PathError{
				Op:   "open",
				Path: "/proc/1234",
				Err:  syscall.ESRCH}, fs.ErrNotExist),
		},
		{
			name:     "Error already including fs.ErrNotExist",
			err:      errors.New("file does not exist"),
			expected: errors.New("file does not exist"),
		},
		{
			name:     "Wrapped error already including fs.ErrNotExist",
			err:      errors.Wrap(fs.ErrNotExist, "wrapped error"),
			expected: errors.Wrap(fs.ErrNotExist, "wrapped error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			flattenErrStrings := func(e error) []string {
				var out []string

				for errStack := []error{e}; len(errStack) > 0; errStack = errStack[1:] {
					err := errStack[0]

					if err == nil {
						continue
					}

					if errs, ok := err.(errors.MultiError); ok {
						errStack = append(errStack, errs.Errors()...)
						continue
					}

					out = append(out, err.Error())
				}

				sort.Strings(out)
				return out
			}

			actualErrs := convertESRCH(tt.err)

			if diff := cmp.Diff(flattenErrStrings(tt.expected), flattenErrStrings(actualErrs)); diff != "" {
				t.Errorf("convertESRCH() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func TestMemoryUsageForPidAndChildren(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx := context.Background()

	var spyRSSCalls []int
	var spyChildrenCalls []int

	// Set up a process tree with PIDs 1: 2, 3
	//                                 3: 4
	// where each process has a memory usage of 2^pid (so that we can easily check the sum of memory usage).
	//
	// Process 2 and 4 will disappear during the call to memoryUsageForPidAndChildren, and we should handle that
	// gracefully.

	proc := &mockProcLike{
		t: t,
		mockRSS: func(t *testing.T, pid int) (uint64, error) {
			spyRSSCalls = append(spyRSSCalls, pid)

			if pid == 2 || pid == 4 {
				// Say that the process has disappeared
				return 0, fs.ErrNotExist
			}

			return uint64(1 << pid), nil
		},
		mockChildren: func(t *testing.T, pid int) ([]int, error) {
			spyChildrenCalls = append(spyChildrenCalls, pid)

			if pid == 1 {
				// Simulate a child process with PIDs 2, 3
				return []int{2, 3}, nil
			}

			if pid == 3 {
				// Simulate a child process with PID 4
				return []int{4}, nil
			}

			return nil, fs.ErrNotExist
		},
	}

	usage, err := memoryUsageForPidAndChildren(ctx, proc, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if diff := cmp.Diff([]int{1, 2, 3, 4}, spyRSSCalls); diff != "" {
		t.Fatalf("RSS calls mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{1, 3}, spyChildrenCalls); diff != "" {
		t.Fatalf("Children calls mismatch (-want +got):\n%s", diff)
	}

	expectedUsage := uint64(1<<1 + 1<<3)
	if usage != expectedUsage {
		t.Errorf("Expected memory usage %d, got %d", expectedUsage, usage)
	}
}

func TestMaxMemoryUsageErrorObserverNotStarted(t *testing.T) {
	defer goleak.VerifyNone(t)

	cmd := allocatingGoProgram(t, 50*1024*1024) // 50 MB
	err := cmd.Start()
	if err != nil {
		t.Fatalf("failed to start test program: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	observer, err := NewLinuxObserver(context.Background(), cmd, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create observer: %v", err)
	}

	_, err = observer.MaxMemoryUsage()
	if !errors.Is(err, errObserverNotStarted) {
		t.Errorf("expected errObserverNotStarted, got: %v", err)
	}
}

type mockProcLike struct {
	t *testing.T

	mockRSS      func(*testing.T, int) (uint64, error)
	mockChildren func(*testing.T, int) ([]int, error)
}

func (m *mockProcLike) RSS(pid int) (uint64, error) {
	if m.mockRSS != nil {
		return m.mockRSS(m.t, pid)
	}

	m.t.Fatal("RSS not implemented")
	return 0, nil
}

func (m *mockProcLike) Children(pid int) ([]int, error) {
	if m.mockChildren != nil {
		return m.mockChildren(m.t, pid)
	}

	m.t.Fatal("Children not implemented")
	return nil, nil
}

func BenchmarkLinuxObservationApproaches(b *testing.B) {
	b.Run("Observer", func(b *testing.B) {
		for _, interval := range []time.Duration{1 * time.Millisecond, 10 * time.Millisecond, 100 * time.Millisecond} {
			b.Run(interval.String(), func(b *testing.B) {
				benchFunc(b, interval)
			})
		}
	})

	b.Run("NoObserver", func(b *testing.B) {
		benchFunc(b, 0)
	})
}

func benchFunc(b *testing.B, observerInterval time.Duration) {
	defer goleak.VerifyNone(b)

	for range b.N {
		workerPool := pool.New().WithErrors()

		for range runtime.NumCPU() {
			workerPool.Go(func() error {
				cmd := allocatingGoProgram(b, 50*1024*1024)

				var buf bytes.Buffer
				cmd.Stderr = &buf
				err := cmd.Start()
				if err != nil {
					return errors.Errorf("starting command: %v, stdErr: %s", err, buf.String())
				}

				observer := NewNoOpObserver()

				if observerInterval > 0 {
					observer, err = NewLinuxObserver(context.Background(), cmd, observerInterval)
					if err != nil {
						return errors.Errorf("failed to create observer: %v", err)
					}
				}

				observer.Start()
				defer observer.Stop()

				err = cmd.Wait()
				if err != nil {
					return errors.Errorf("waiting for command: %v, stdErr: %s", err, buf.String())
				}

				_, isNoOpObserver := observer.(*noopObserver)
				if isNoOpObserver {
					return nil
				}

				memory, err := observer.MaxMemoryUsage()
				if err != nil {
					return errors.Errorf("getting memory usage: %v", err)
				}

				memoryLow := bytesize.Size(10 << 20)   // 10MB
				memoryHigh := bytesize.Size(100 << 20) // 100MB

				if !(memoryLow < memory && memory < memoryHigh) {
					return errors.Errorf("memory usage is not in the expected range (low: %s, high: %s): %s", humanize.Bytes(uint64(memoryLow)), humanize.Bytes(uint64(memoryHigh)), humanize.Bytes(uint64(memory)))
				}

				return nil
			})

			if err := workerPool.Wait(); err != nil {
				b.Fatalf("error in worker pool: %v", err)
			}
		}
	}
}

var _ processInfoProvider = &mockProcLike{}
