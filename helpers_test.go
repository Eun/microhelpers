package microhelpers

import (
	"net/http"
	"os"
	"testing"

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

func TestParseEnvironment(t *testing.T) {
	t.Run("Nothing", func(t *testing.T) {
		_, _, err := Parse(nil, http.NewServeMux())
		require.Error(t, err)
	})
	t.Run("Port", func(t *testing.T) {
		os.Setenv("APP_PORT", "8000")
		defer os.Unsetenv("APP_PORT")
		addresses, _, err := Parse(nil, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)
	})
	t.Run("Address", func(t *testing.T) {
		os.Setenv("APP_ADDRESS", ":8000")
		defer os.Unsetenv("APP_ADDRESS")
		addresses, _, err := Parse(nil, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)
	})
	t.Run("AddressAndPort", func(t *testing.T) {
		os.Setenv("APP_ADDRESS", "127.0.0.1")
		os.Setenv("APP_PORT", "8000")
		defer os.Unsetenv("APP_ADDRESS")
		defer os.Unsetenv("APP_PORT")
		addresses, _, err := Parse(nil, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{"127.0.0.1:8000"}, addresses)
	})
	t.Run("AddressAndPort-Override", func(t *testing.T) {
		os.Setenv("APP_ADDRESS", "127.0.0.1:8001")
		os.Setenv("APP_PORT", "8000")
		defer os.Unsetenv("APP_ADDRESS")
		defer os.Unsetenv("APP_PORT")
		addresses, _, err := Parse(nil, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{"127.0.0.1:8001"}, addresses)
	})
	t.Run("Root", func(t *testing.T) {
		os.Setenv("APP_ROOT", "/subdir/")
		os.Setenv("APP_PORT", "8000")
		defer os.Unsetenv("APP_ROOT")
		defer os.Unsetenv("APP_PORT")

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(1234)
		})

		addresses, handler, err := Parse(nil, mux)
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)

		var buf mockResponseWriter
		r, err := http.NewRequest(http.MethodGet, "/subdir/", nil)
		require.NoError(t, err)
		handler.ServeHTTP(&buf, r)
		require.Equal(t, 1234, buf.StatusCode)

		buf = mockResponseWriter{}
		r, err = http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		handler.ServeHTTP(&buf, r)
		require.Equal(t, 404, buf.StatusCode)
	})
}

func TestParseCommandline(t *testing.T) {
	t.Run("Nothing", func(t *testing.T) {
		_, _, err := Parse(nil, http.NewServeMux())
		require.Error(t, err)
	})
	t.Run("Port", func(t *testing.T) {
		addresses, _, err := Parse([]string{"--port", "8000"}, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)
	})
	t.Run("Address", func(t *testing.T) {
		addresses, _, err := Parse([]string{"--address", ":8000"}, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)
	})
	t.Run("AddressAndPort", func(t *testing.T) {
		addresses, _, err := Parse([]string{"--port", "8000", "--address", "127.0.0.1"}, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{"127.0.0.1:8000"}, addresses)
	})
	t.Run("AddressAndPort-Override", func(t *testing.T) {
		addresses, _, err := Parse([]string{"--port", "8000", "--address", "127.0.0.1:8001"}, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{"127.0.0.1:8001"}, addresses)
	})
	t.Run("Root", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(1234)
		})

		addresses, handler, err := Parse([]string{"--port", "8000", "--root", "/subdir/"}, mux)
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)

		var buf mockResponseWriter
		r, err := http.NewRequest(http.MethodGet, "/subdir/", nil)
		require.NoError(t, err)
		handler.ServeHTTP(&buf, r)
		require.Equal(t, 1234, buf.StatusCode)

		buf = mockResponseWriter{}
		r, err = http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		handler.ServeHTTP(&buf, r)
		require.Equal(t, 404, buf.StatusCode)
	})
	t.Run("CommandLineOverridesEnvironment", func(t *testing.T) {
		os.Setenv("APP_PORT", "6000")
		defer os.Unsetenv("APP_PORT")
		addresses, _, err := Parse([]string{"--port", "8000"}, http.NewServeMux())
		require.NoError(t, err)
		require.EqualValues(t, []string{":8000"}, addresses)
	})
}
