package myhttp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	nhttp "net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

const (
	HeaderAccept         = "Accept"
	HeaderContentType    = "Content-Type"
	HeaderAAuthorization = "Authorization"
	ContentTypeJson      = "application/json"
	Scheme               = "https://"
)

// MakeHttpClient returns client ready to make HTTP requests.
// It's primed with certs loaded from the given caPath.
// If no caPath provided, TLS will be unauthenticated.
// The certs are used to establish that the servers are who they say they are.
func MakeHttpClient(caPath string) (*nhttp.Client, error) {
	pool, err := loadCertPoolFromFile(caPath)
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(nil /* no options */)
	if err != nil {
		return nil, err
	}
	return &nhttp.Client{
		Transport: makeTransport(makeTlsConfig(pool)),
		Timeout:   8 * time.Second,
		// Don't automatically follow redirects; we want debug mode to
		// expose redirect hops.
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//	return http.ErrUseLastResponse
		// },
		Jar: jar,
	}, nil
}

// loadCertPoolFromFile returns a pool containing the certs read from the given file.
func loadCertPoolFromFile(path string) (*x509.CertPool, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load certs from %q; %w", path, err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(data)
	return pool, nil
}

func makeTransport(tlsConfig *tls.Config) *nhttp.Transport {
	return &nhttp.Transport{
		TLSClientConfig:     tlsConfig,
		Proxy:               nhttp.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// makeTlsConfig returns a TLS config that uses a cert pool if provided, falling
// back to no cert check if no cert pool provided.
func makeTlsConfig(pool *x509.CertPool) *tls.Config {
	if pool == nil {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return &tls.Config{
		RootCAs: pool,
	}
}

type PrArgs struct {
	Headers bool
	Body    bool
}

// PrintRequest prints a request.
func PrintRequest(r *nhttp.Request, args PrArgs) error {
	doErrF("==== %s request to %s\n", r.Method, r.URL)
	if args.Headers {
		printHeaders(r.Header)
	}
	if args.Body {
		if err := dumpIt("REQUEST", r.Body); err != nil {
			return err
		}
	}
	return nil
}

// PrintResponse prints a response.
func PrintResponse(r *nhttp.Response, args PrArgs) error {
	doErrF("==== Response Code %d ==================\n", r.StatusCode)
	if args.Headers {
		printHeaders(r.Header)
	}
	if args.Body {
		if err := dumpIt("RESPONSE", r.Body); err != nil {
			return err
		}
	}
	return nil
}

func dumpIt(title string, body io.ReadCloser) error {
	const bodyDelim = " -------------------------------"
	b, err := io.ReadAll(body)
	if err != nil {
		// Doing this makes the body unavailable elsewhere!
		_ = body.Close()
		return err
	}
	doErr1("BEGIN", title, bodyDelim)
	if body := strings.TrimSpace(string(b)); body != "" {
		doErr1(body)
	}
	doErr1("END", title, bodyDelim)
	// Doing this makes the body unavailable elsewhere!
	_ = body.Close()
	return nil
}

// printHeaders prints response headers.
func printHeaders(m map[string][]string) {
	for k, v := range m {
		doErrF("%50s : %s\n", k, v)
	}
}

func doErrF(f string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, f, args...)
}

func doErr1(f ...string) {
	_, _ = fmt.Fprintln(os.Stderr, f)
}
