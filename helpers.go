package microhelpers

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
)

type Settings struct {
	// Address is the listening address
	Address []string
	// Port is the listening port
	Port []string
	// Root is the root path of your handler
	Root []string
}
type Config struct {
	Environment Settings
	CommandLine Settings
}

func Parse(args []string, handler http.Handler, config ...*Config) ([]string, http.Handler, error) {
	if handler == nil {
		return nil, nil, errors.New("No handler specified")
	}
	app := kingpin.New("", "")

	var cfg *Config

	for i := 0; i < len(config); i++ {
		if config[i] != nil {
			cfg = config[i]
			break
		}
	}

	if cfg == nil {
		cfg = &Config{
			Environment: Settings{
				Port:    []string{"PORT", "APP_PORT", "HTTP_PLATFORM_PORT", "ASPNETCORE_PORT"},
				Address: []string{"ADDRESS", "APP_ADDRESS"},
				Root:    []string{"APP_ROOT"},
			},
			CommandLine: Settings{
				Port:    []string{"port"},
				Address: []string{"addr", "address"},
				Root:    []string{"root"},
			},
		}
	}

	var port uint32
	var addresses []string
	var root string
	// parse port from environment
	for _, s := range cfg.Environment.Port {
		if env := os.Getenv(s); len(env) > 0 {
			i, err := strconv.ParseUint(env, 10, 16)
			if err == nil {
				port = uint32(i)
				break
			}
		}
	}

	// parse addr from environment
	for _, s := range cfg.Environment.Address {
		if env := os.Getenv(s); len(env) > 0 {
			addresses = append(addresses, env)
			break
		}
	}

	// parse root from environment
	for _, s := range cfg.Environment.Root {
		if env := os.Getenv(s); len(env) > 0 {
			root = env
			break
		}
	}

	// parse port from commandline
	portFlags := make([]*uint32, 0, len(cfg.CommandLine.Port))
	for _, s := range cfg.CommandLine.Port {
		portFlags = append(portFlags, app.Flag(s, "").Uint32())
	}

	// parse address from commandline
	addressFlags := make([]*[]string, 0, len(cfg.CommandLine.Address))
	for _, s := range cfg.CommandLine.Address {
		addressFlags = append(addressFlags, app.Flag(s, "").Strings())
	}

	// parse root from commandline
	rootFlags := make([]*string, 0, len(cfg.CommandLine.Root))
	for _, s := range cfg.CommandLine.Root {
		rootFlags = append(rootFlags, app.Flag(s, "").String())
	}

	_, err := app.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	for _, portFlag := range portFlags {
		if portFlag != nil && *portFlag > 0 && *portFlag < 65536 {
			port = *portFlag
			break
		}
	}

	for _, addressFlag := range addressFlags {
		if addressFlag != nil && len(*addressFlag) > 0 {
			for _, address := range *addressFlag {
				addresses = append(addresses, address)
			}
		}
	}

	for _, rootFlag := range rootFlags {
		if rootFlag != nil && len(*rootFlag) > 0 {
			root = *rootFlag
			break
		}
	}

	root = "/" + strings.Trim(filepath.ToSlash(root), "/")
	if root != "/" {
		mux := http.NewServeMux()
		mux.Handle(root, http.RedirectHandler(root+"/", http.StatusTemporaryRedirect))
		mux.Handle(root+"/", http.StripPrefix(root, handler))
		handler = mux
	}

	if len(addresses) == 0 {
		addresses = append(addresses, "")
	}

	for i, address := range addresses {
		if _, _, err := net.SplitHostPort(address); err != nil {
			if port == 0 {
				return nil, nil, fmt.Errorf("Unable to find port, make sure to use one of the command line switches (%s) or one of the environment variables (%s)", strings.Join(cfg.CommandLine.Port, ", "), strings.Join(cfg.Environment.Port, ", "))
			}
			addresses[i] = fmt.Sprintf("%s:%d", address, port)
		}
	}

	return addresses, handler, nil
}

func ListenAndServe(addresses []string, handler http.Handler, logger ...io.Writer) error {
	size := len(addresses)
	if size <= 0 {
		return errors.New("No addresses to listen on")
	}

	var firstLogger io.Writer

	for i := 0; i < len(logger); i++ {
		if logger[i] != nil {
			firstLogger = logger[i]
			break
		}
	}

	errChan := make(chan error, size)

	for i := 0; i < size; i++ {
		go func(address string) {
			if firstLogger != nil {
				fmt.Fprintf(firstLogger, "Listening on %s\n", address)
			}
			if err := http.ListenAndServe(address, handler); err != nil {
				errChan <- err
			}
		}(addresses[i])
	}

	return <-errChan
}
