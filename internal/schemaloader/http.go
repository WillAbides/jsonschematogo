package schemaloader

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

func tlsConfigWithCACert(tlsConfig *tls.Config, cacert string) (*tls.Config, error) {
	tlsConfig = tlsConfig.Clone()
	if cacert == "" {
		return tlsConfig, nil
	}
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	if tlsConfig.RootCAs == nil {
		tlsConfig.RootCAs = x509.NewCertPool()
	}
	pem, err := os.ReadFile(cacert)
	if err != nil {
		return nil, err
	}
	if !tlsConfig.RootCAs.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("appending CA certs from %s", cacert)
	}
	return tlsConfig, nil
}

func tlsConfigWithInsecure(tlsConfig *tls.Config, insecure bool) *tls.Config {
	tlsConfig = tlsConfig.Clone()
	if !insecure {
		return tlsConfig
	}
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	tlsConfig.InsecureSkipVerify = true
	return tlsConfig
}

func setupTlsConfig(transport *http.Transport, insecure bool, cacert string) error {
	tlsConfig, err := tlsConfigWithCACert(transport.TLSClientConfig, cacert)
	if err != nil {
		return fmt.Errorf("setting up TLS config: %w", err)
	}
	transport.TLSClientConfig = tlsConfigWithInsecure(tlsConfig, insecure)
	return nil
}

func newHTTPClient(insecure bool, cacert string) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	err := setupTlsConfig(transport, insecure, cacert)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: transport,
		Timeout:   15 * 1000, // 15 seconds
	}, nil
}
