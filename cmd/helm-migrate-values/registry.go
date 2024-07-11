package main

import (
	"helm.sh/helm/v3/pkg/registry"
	"os"
)

func newRegistryClient(certFile, keyFile, caFile string, insecureSkipTLSverify, plainHTTP bool) (*registry.Client, error) {
	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSverify {
		registryClient, err := newRegistryClientWithTLS(certFile, keyFile, caFile, insecureSkipTLSverify)
		if err != nil {
			return nil, err
		}
		return registryClient, nil
	}
	registryClient, err := newDefaultRegistryClient(plainHTTP)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newDefaultRegistryClient(plainHTTP bool) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}
	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClientWithTLS(certFile, keyFile, caFile string, insecureSkipTLSverify bool) (*registry.Client, error) {
	// Create a new registry client
	registryClient, err := registry.NewRegistryClientWithTLS(os.Stderr, certFile, keyFile, caFile, insecureSkipTLSverify,
		settings.RegistryConfig, settings.Debug,
	)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}
