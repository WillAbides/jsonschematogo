package schemaloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

var _ jsonschema.URLLoader = (*URLLoader)(nil)

type OnLoadFunc func(url string, schema any)

type URLLoader struct {
	mappingsLoader mappingsLoader
	onLoad         OnLoadFunc
}

func (s *URLLoader) Load(url string) (any, error) {
	schema, err := s.mappingsLoader.Load(url)
	if err != nil {
		return nil, fmt.Errorf("loading schema %q: %w", url, err)
	}
	if s.onLoad != nil {
		s.onLoad(url, schema)
	}
	return schema, nil
}

type Options struct {
	Mappings map[string]string
	Insecure bool
	CACert   string
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
				"http":  (*httpLoader)(httpClient),
				"https": (*httpLoader)(httpClient),
			},
		},
	}, nil
}

type loaderFunc func(string) (any, error)

func (f loaderFunc) Load(url string) (any, error) { return f(url) }

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

func loadFile(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadBytes(data)
}

type httpLoader http.Client

func (l *httpLoader) Load(url string) (_ any, errOut error) {
	client := (*http.Client)(l)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { errOut = errors.Join(resp.Body.Close()) }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status code %d", url, resp.StatusCode)
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

func (l *mappingsLoader) Load(url string) (any, error) {
	for prefix, dir := range l.mappings {
		if suffix, ok := strings.CutPrefix(url, prefix); ok {
			return loadFile(filepath.Join(dir, suffix))
		}
	}
	return l.fallback.Load(url)
}
