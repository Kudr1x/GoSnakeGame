package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig_ParseFlags(t *testing.T) {
	t.Parallel()

	cfg := DefaultServerConfig()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlags(fs)

	err := fs.Parse([]string{"-addr", "localhost:1234", "-width", "30", "-height", "30"})
	assert.NoError(t, err)

	assert.Equal(t, "localhost:1234", cfg.Addr)
	assert.Equal(t, 30, cfg.Width)
	assert.Equal(t, 30, cfg.Height)
}

func TestClientConfig_ParseFlags(t *testing.T) {
	t.Parallel()

	cfg := DefaultClientConfig()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlags(fs)

	err := fs.Parse([]string{"-server", "remotehost:5678"})
	assert.NoError(t, err)

	assert.Equal(t, "remotehost:5678", cfg.ServerAddr)
}
