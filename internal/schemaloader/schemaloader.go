package schemaloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

var _ jsonschema.URLLoader = (*URLLoader)(nil)

type OnLoadFunc func(uri string, schema any)

type URLLoader struct {
	mappingsLoader mappingsLoader
	onLoad         OnLoadFunc
}

func (s *URLLoader) Load(u string) (any, error) {
	schema, err := s.mappingsLoader.Load(u)
	if err != nil {
		return nil, err
	}
	if s.onLoad != nil {
		s.onLoad(u, schema)
	}
	return schema, nil
}

type Options struct {
	Mappings map[string]string
	CACert   string
	Insecure bool
}

func New(onLoad OnLoadFunc, opts *Options) (*URLLoader, error) {
	if opts == nil {
		opts = &Options{}
	}
	httpClient, err := newHTTPClient(opts.Insecure, opts.CACert)
	if err != nil {
		return nil, err
	}
	return &URLLoader{
		onLoad: onLoad,
		mappingsLoader: mappingsLoader{
			mappings: opts.Mappings,
			fallback: jsonschema.SchemeURLLoader{
				"file":  loaderFunc(loadFile),
				"":      loaderFunc(loadFile),
				"http":  (*httpLoader)(httpClient),
				"https": (*httpLoader)(httpClient),
			},
		},
	}, nil
}

type loaderFunc func(string) (any, error)

func (f loaderFunc) Load(u string) (any, error) { return f(u) }

func loadBytes(data []byte) (any, error) {
	if json.Valid(data) {
		return jsonschema.UnmarshalJSON(bytes.NewReader(data))
	}
	var v any
	err := yaml.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func loadFile(u string) (any, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("parsing URL %q: %w", u, err)
	}
	if parsedURL.Scheme != "file" && parsedURL.Scheme != "" {
		return nil, fmt.Errorf("unsupported URL scheme %q for file loading", parsedURL.Scheme)
	}

	filename := parsedURL.Path
	if strings.HasPrefix(filename, "/") && os.PathSeparator != '/' {
		// Convert absolute path to OS-specific path
		filename = strings.TrimPrefix(filename, "/")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return loadBytes(data)
}

type httpLoader http.Client

func (l *httpLoader) Load(u string) (_ any, errOut error) {
	client := (*http.Client)(l)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { errOut = errors.Join(resp.Body.Close()) }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status code %d", u, resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return loadBytes(b)
}

type mappingsLoader struct {
	mappings map[string]string
	fallback jsonschema.URLLoader
}

func (l *mappingsLoader) Load(u string) (any, error) {
	for prefix, dir := range l.mappings {
		suffix, ok := strings.CutPrefix(u, prefix)
		if ok {
			return loadFile(filepath.Join(dir, suffix))
		}
	}
	return l.fallback.Load(u)
}
