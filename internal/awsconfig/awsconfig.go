// Package awsconfig generates AWS client configurations optimized for use by
// the randomizer.
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
	"github.com/pkg/errors"
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
func New(ctx context.Context, extraOptions ...Option) (aws.Config, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// This option is recommended in AWS Lambda deployments due to the
	// significant reduction in cold start latency (see getEmbeddedCertPool).
	// It can be enabled for standard server deployments if desired, but is
	// likely far less beneficial.
	if os.Getenv("AWS_CLIENT_EMBEDDED_TLS_ROOTS") == "1" {
		transport.TLSClientConfig = &tls.Config{
			RootCAs: getEmbeddedCertPool(),
		}
	}

	client := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: transport,
	}

	options := []Option{
		config.WithHTTPClient(client),
		config.WithRetryer(
			func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), DefaultRetryMaxAttempts)
			},
		),
	}
	options = append(options, extraOptions...)

	cfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return aws.Config{}, errors.Wrap(err, "loading AWS config")
	}

	// WARNING: X-Ray tracing will panic if the context passed to AWS operations
	// is not already associated with an open X-Ray segment. That means that as of
	// this writing this option is only safe to use on AWS Lambda. Standard server
	// deployments should avoid setting it.
	if useXRay := os.Getenv("AWS_CLIENT_XRAY_TRACING"); useXRay == "1" {
		xrayawsv2.AWSV2Instrumentor(&cfg.APIOptions)
	}

	return cfg, errors.Wrap(err, "loading AWS config")
}

//go:generate ./refresh-amazon-trust-roots.sh
//go:embed amazon-trust.cer
var embeddedRootsDER []byte

// getEmbeddedCertPool returns a cert pool containing only the root CAs
// operated by Amazon Trust Services, which all AWS service endpoints chain
// from.
//
// When the randomizer runs on AWS Lambda in the recommended configuration,
// this limited set of roots is so much cheaper to parse than the full set of
// roots trusted by Amazon Linux 2 that it cuts invocation time on cold starts
// approximately in half, specifically by around 500ms. This is a large enough
// difference for a human to notice, and accounts for about 15% of the 3 second
// response time limit that Slack imposes on slash commands.
var getEmbeddedCertPool = sync.OnceValue(func() *x509.CertPool {
	certs, err := x509.ParseCertificates(embeddedRootsDER)
	if err != nil {
		panic(fmt.Errorf("failed to parse embedded TLS roots: %v", err))
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool
})
