package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"roblox-sync/sync-agent/internal/mapfs"
	"roblox-sync/sync-agent/internal/serve"
)

const (
	AgentVersion    = "0.1.0"
	ProtocolVersion = 1
	defaultBind     = "127.0.0.1"
	defaultPort     = 29850
	defaultProject  = "."
	minPort         = 1024
	maxPort         = 65535
	forbiddenPort   = 1194
	tokenBytes      = 32
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(args) == 0 {
		return help(stdout)
	}
	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "agent=%s protocol=%d\n", AgentVersion, ProtocolVersion)
		return 0
	case "help":
		return help(stdout)
	case "serve":
		return runServe(ctx, args[1:], stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}

func help(stdout io.Writer) int {
	fmt.Fprint(stdout, `Usage: rbx-sync <command>
Commands:
  serve    Load project, bind, print token
  version  Print agent and protocol versions
  help     Show usage
`)
	return 0
}

func runServe(ctx context.Context, args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	project := fs.String("project", defaultProject, "directory containing rbx-sync.json")
	bind := fs.String("bind", defaultBind, "loopback bind host")
	port := fs.Int("port", defaultPort, "listen port")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *bind != "127.0.0.1" && *bind != "::1" {
		fmt.Fprintln(stderr, "must be loopback")
		return 1
	}
	if *port < minPort || *port > maxPort || *port == forbiddenPort {
		fmt.Fprintf(stderr, "port must be in %d–%d excluding %d\n", minPort, maxPort, forbiddenPort)
		return 1
	}
	fsys, err := mapfs.Open(*project)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(*bind, strconv.Itoa(*port)))
	if err != nil {
		if addrInUse(err) {
			fmt.Fprintln(stderr, "address in use")
		} else {
			fmt.Fprintln(stderr, err.Error())
		}
		return 1
	}
	token, err := newAccessToken()
	if err != nil {
		ln.Close()
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	httpSrv := &http.Server{
		Handler: serve.New(fsys, token, AgentVersion, ProtocolVersion).Handler(),
	}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpSrv.Serve(ln)
	}()
	fmt.Fprintf(stderr, "protocolVersion=%d\n", ProtocolVersion)
	fmt.Fprintf(stderr, "listen=%s\n", ln.Addr().String())
	fmt.Fprintf(stderr, "accessToken=%s\n", token)
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		<-serveErr
		return 0
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	}
}

func newAccessToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func addrInUse(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "address already in use") ||
		strings.Contains(s, "only one usage of each socket address") ||
		strings.Contains(s, "address in use")
}
