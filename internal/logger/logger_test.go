package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		development bool
		wantErr     bool
	}{
		{
			name:        "production logger",
			development: false,
			wantErr:     false,
		},
		{
			name:        "development logger",
			development: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Init(tt.development)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				logger := Get()
				if logger == nil {
					t.Error("Get() returned nil logger")
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	// Test getting logger before init
	logger := Get()
	if logger == nil {
		t.Error("Get() should return non-nil logger even before Init()")
	}
}

func TestSync(t *testing.T) {
	t.Parallel()

	// Test sync before init
	err := Sync()
	if err != nil {
		t.Errorf("Sync() before Init() should not error, got: %v", err)
	}

	// Test sync after init - ignore stderr sync errors in tests
	if err := Init(false); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Sync() may fail on stderr in test environment, which is expected
	_ = Sync()
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	if err := Init(false); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			logger := Get()
			logger.Info("test message", zap.Int("id", i))
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
