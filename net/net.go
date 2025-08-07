//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func RequestBody(r *http.Request) (bod []byte, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err = errors.Join(err, b.Close())
	}(r.Body)

	return body, err
}

// CreateMtlsServer creates an HTTP server configured for mutual TLS (mTLS)
// authentication using SPIFFE X.509 certificates. It sets up the server with a
// custom authorizer that validates client SPIFFE IDs against a provided
// predicate function.
//
// Parameters:
//   - source: An X509Source that provides the server's identity credentials and
//     validates client certificates. It must be initialized and valid.
//   - tlsPort: The network address and port for the server to listen on
//     (e.g., ":8443").
//   - predicate: A function that takes a SPIFFE ID string and returns true if
//     the client should be allowed access, false otherwise.
//
// Returns:
//   - *http.Server: A configured HTTP server ready to be started with TLS
//     enabled.
//   - error: An error if the server configuration fails.
//
// The server uses the provided X509Source for both its own identity and for
// validating client certificates. Client connections are only accepted if their
// SPIFFE ID passes the provided predicate function.
func CreateMtlsServer(source *workloadapi.X509Source,
	tlsPort string,
	predicate func(string) bool) (*http.Server, error) {
	authorizer := tlsconfig.AdaptMatcher(func(id spiffeid.ID) error {
		if predicate(id.String()) {
			return nil
		}

		return fmt.Errorf(
			"authorizer: TLS Config: untrusted spiffe id: '%s'", id.String(),
		)
	})

	tlsConfig := tlsconfig.MTLSServerConfig(source, source, authorizer)
	server := &http.Server{
		Addr:      tlsPort,
		TLSConfig: tlsConfig,
	}
	return server, nil
}

// CreateMtlsClientWithPredicate creates an HTTP client configured for mutual TLS
// authentication using SPIFFE workload identities.
// It uses the provided X.509 source for client certificates and validates peer
// certificates against a predicate function.
//
// Parameters:
//   - source: An X509Source that provides the client's identity certificates
//     and trusted roots
//   - predicate: A function that evaluates SPIFFE IDs (as strings) and returns
//     true if the ID should be trusted
//
// Returns:
//   - *http.Client: A configured HTTP client that will use mTLS for all
//     connections
//   - error: An error if the client creation fails
//
// The returned client will:
//   - Present client certificates from the provided X509Source
//   - Validate peer certificates using the same X509Source
//   - Only accept peer certificates with SPIFFE IDs that pass the predicate
//     function
func CreateMtlsClientWithPredicate(
	source *workloadapi.X509Source,
	predicate func(string) bool,
) (*http.Client, error) {
	authorizer := tlsconfig.AdaptMatcher(func(id spiffeid.ID) error {
		if predicate(id.String()) {
			return nil
		}

		return fmt.Errorf(
			"TLS Config: untrusted spiffe id: '%s'", id.String(),
		)
	})

	tlsConfig := tlsconfig.MTLSClientConfig(source, source, authorizer)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			IdleConnTimeout:     30 * time.Second,
			MaxIdleConns:        100,
			MaxConnsPerHost:     10,
			MaxIdleConnsPerHost: 10,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	return client, nil
}

// CreateMtlsClient creates an HTTP client configured for mutual TLS
// authentication using SPIFFE workload identities.
// It uses the provided X.509 source for client certificates and validates peer
// certificates against a predicate function.
//
// Parameters:
//   - source: An X509Source that provides the client's identity certificates
//     and trusted roots
//
// Returns:
//   - *http.Client: A configured HTTP client that will use mTLS for all
//     connections
//   - error: An error if the client creation fails
//
// The returned client will:
//   - Present client certificates from the provided X509Source
//   - Validate peer certificates using the same X509Source
//   - Only accept peer certificates with SPIFFE IDs that pass the predicate
//     function
func CreateMtlsClient(source *workloadapi.X509Source) (*http.Client, error) {
	return CreateMtlsClientWithPredicate(source, func(string) bool { return true })
}

// ServeWithPredicate initializes and starts an HTTPS server using mTLS
// authentication with SPIFFE X.509 certificates. It sets up the server routes
// using the provided initialization function and listens for incoming
// connections on the specified port.
//
// Parameters:
//   - source: An X509Source that provides the server's identity credentials and
//     validates client certificates. Must not be nil.
//   - initializeRoutes: A function that sets up the HTTP route handlers for the
//     server. This function is called before the server starts.
//   - predicate: a predicate function to pass to CreateMtlsServer.
//   - tlsPort: The network address and port for the server to listen on
//     (e.g., ":8443").
//
// Returns:
//   - error: Returns nil if the server starts successfully, otherwise returns
//     an error explaining the failure. Specific error cases include:
//   - If source is nil
//   - If server creation fails
//   - If the server fails to start or encounters an error while running
//
// The function uses empty strings for the certificate and key file parameters
// in ListenAndServeTLS as the certificates are provided by the X509Source. The
// server's mTLS configuration is determined by the CreateMtlsServer function.
func ServeWithPredicate(source *workloadapi.X509Source,
	initializeRoutes func(),
	predicate func(string) bool,
	tlsPort string) error {
	if source == nil {
		return errors.New("serve: got nil source while trying to serve")
	}

	initializeRoutes()

	server, err := CreateMtlsServer(source, tlsPort, predicate)
	if err != nil {
		return err
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		return errors.Join(
			err,
			errors.New("serve: failed to listen and serve"),
		)
	}

	return nil
}

// Serve initializes and starts an HTTPS server using mTLS
// authentication with SPIFFE X.509 certificates. It sets up the server routes
// using the provided initialization function and listens for incoming
// connections on the specified port.
//
// Parameters:
//   - source: An X509Source that provides the server's identity credentials and
//     validates client certificates. Must not be nil.
//   - initializeRoutes: A function that sets up the HTTP route handlers for the
//     server. This function is called before the server starts.
//   - tlsPort: The network address and port for the server to listen on
//     (e.g., ":8443").
//
// Returns:
//   - error: Returns nil if the server starts successfully, otherwise returns
//     an error explaining the failure. Specific error cases include:
//   - If source is nil
//   - If server creation fails
//   - If the server fails to start or encounters an error while running
//
// The function uses empty strings for the certificate and key file parameters
// in ListenAndServeTLS as the certificates are provided by the X509Source. The
// server's mTLS configuration is determined by the CreateMtlsServer function.
func Serve(
	source *workloadapi.X509Source,
	initializeRoutes func(),
	tlsPort string) error {
	return ServeWithPredicate(
		source, initializeRoutes,
		func(string) bool { return true }, tlsPort)
}
