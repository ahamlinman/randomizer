// Package awsconfig provides a randomizer-optimized AWS SDK configuration.
package awsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	xrayawsv2 "github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

const (
	// DefaultTimeout is set to half of the 3-second response time limit that
	// Slack imposes on slash commands.
	DefaultTimeout = 1500 * time.Millisecond

	// DefaultRetryMaxAttempts allows up to 2 attempts to make AWS API calls.
	// Based on Slack's 3-second response time limit and our default timeout,
	// it's unlikely that we'll get many more attempts than this.
	DefaultRetryMaxAttempts = 2
)

type Option = func(*config.LoadOptions) error

// New creates a new AWS client configuration using reasonable default settings
// for timeouts and retries.
func New(ctx context.Context) (aws.Config, error) {
	transport := http.DefaultTransport

	// This option is recommended in AWS Lambda to significantly reduce cold
	// start latency (see [getEmbeddedCertTransport]). It can be enabled for
	// standard server deployments if desired, but is far less beneficial.
	if os.Getenv("AWS_CLIENT_EMBEDDED_TLS_ROOTS") == "1" {
		transport = getEmbeddedCertTransport()
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithHTTPClient(&http.Client{
			Timeout:   DefaultTimeout,
			Transport: transport,
		}),
		config.WithRetryer(
			func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), DefaultRetryMaxAttempts)
			},
		),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config: %w", err)
	}

	// NOTE: X-Ray tracing panics if the context for an AWS call is not already
	// associated with an open X-Ray segment. As of writing, this option is only
	// safe to use on AWS Lambda. Standard server deployments should avoid it.
	if useXRay := os.Getenv("AWS_CLIENT_XRAY_TRACING"); useXRay == "1" {
		xrayawsv2.AWSV2Instrumentor(&cfg.APIOptions)
	}

	return cfg, nil
}

// getEmbeddedCertTransport returns an HTTP transport that trusts only the root
// CAs operated by Amazon Trust Services, which all AWS service endpoints chain
// from.
//
// When the randomizer runs on AWS Lambda with recommended resource settings,
// this limited set of roots is substantially cheaper to parse than a typical
// root store, which removes ~500ms of cold-start response latency. That's
// large enough for a human to notice, and accounts for ~15% of the 3-second
// response time limit Slack imposes on slash commands.
var getEmbeddedCertTransport = sync.OnceValue(func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: loadEmbeddedCertPool()}
	return transport
})

//go:generate ./refresh-amazon-trust-roots.sh
//go:embed amazon-trust.cer
var embeddedRootsDER []byte

func loadEmbeddedCertPool() *x509.CertPool {
	certs, err := x509.ParseCertificates(embeddedRootsDER)
	if err != nil {
		panic(fmt.Errorf("failed to parse embedded TLS roots: %v", err))
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool
}
