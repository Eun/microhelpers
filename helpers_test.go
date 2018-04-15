package microhelpers

import (
	"net/http"
	"testing"

	"os"

	"github.com/stretchr/testify/require"
)

type mockResponseWriter struct {
	StatusCode int
}

func (*mockResponseWriter) Header() http.Header {
	return http.Header(map[string][]string{})
}
func (*mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func TestParse(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		val, err := ParseString([]string{"config"}, []string{"config"}, "default.json", nil)
		require.NoError(t, err)
		require.Equal(t, "default.json", val)
	})

	t.Run("Flag", func(t *testing.T) {
		t.Run("Without Equal", func(t *testing.T) {
			val, err := ParseString([]string{"config"}, []string{"config"}, "default.json", []string{"--config", "config.json"})
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
		})
		t.Run("With Equal", func(t *testing.T) {
			val, err := ParseString([]string{"config"}, []string{"config"}, "default.json", []string{"--config=config.json"})
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
		})
		t.Run("Second Name", func(t *testing.T) {
			val, err := ParseString([]string{"config", "cfg"}, []string{"config"}, "default.json", []string{"--cfg=config.json"})
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
		})
	})

	t.Run("ShortFlag", func(t *testing.T) {
		t.Run("Without Equal", func(t *testing.T) {
			val, err := ParseString([]string{"c"}, []string{"config"}, "default.json", []string{"-c", "config.json"})
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
		})
		t.Run("With Equal", func(t *testing.T) {
			val, err := ParseString([]string{"c"}, []string{"config"}, "default.json", []string{"-c=config.json"})
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
		})
	})

	t.Run("Env", func(t *testing.T) {
		t.Run("First Name", func(t *testing.T) {
			os.Setenv("CONFIG", "config.json")
			val, err := ParseString([]string{"config"}, []string{"config"}, "default.json", nil)
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
			os.Unsetenv("CONFIG")
		})
		t.Run("Second Name", func(t *testing.T) {
			os.Setenv("CFG", "config.json")
			val, err := ParseString(nil, []string{"config", "cfg"}, "default.json", nil)
			require.NoError(t, err)
			require.Equal(t, "config.json", val)
			os.Unsetenv("CFG")
		})
	})

	t.Run("EnvAndFlag", func(t *testing.T) {
		os.Setenv("CONFIG", "env.json")
		val, err := ParseString([]string{"config"}, []string{"config"}, "default.json", []string{"--config", "flag.json"})
		require.NoError(t, err)
		require.Equal(t, "flag.json", val)
		os.Unsetenv("CONFIG")
	})
}
