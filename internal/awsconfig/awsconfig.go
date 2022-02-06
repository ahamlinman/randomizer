package awsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
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
	DefaultTimeout          = 1500 * time.Millisecond
	DefaultRetryMaxAttempts = 2
)

type Option = func(*config.LoadOptions) error

// New creates a new AWS client configuration using pinned TLS roots and
// reasonable default settings for timeouts and retries.
func New(ctx context.Context, extraOptions ...Option) (aws.Config, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: getTLSRootPool()}

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
	// this writing this option is only safe to use on AWS Lambda. Other clients
	// should avoid setting it.
	if useXRay := os.Getenv("AWS_CLIENT_XRAY_TRACING"); useXRay == "1" {
		xrayawsv2.AWSV2Instrumentor(&cfg.APIOptions)
	}

	return cfg, errors.Wrap(err, "loading AWS config")
}

//go:generate ./refresh-tls-roots.sh

var (
	//go:embed cert.pem
	tlsRootsPEM  []byte
	tlsRoots     *x509.CertPool
	initTLSRoots sync.Once
)

func getTLSRootPool() *x509.CertPool {
	initTLSRoots.Do(func() {
		tlsRoots = x509.NewCertPool()
		if !tlsRoots.AppendCertsFromPEM(tlsRootsPEM) {
			panic("failed to initialize TLS roots for DynamoDB client")
		}
	})
	return tlsRoots
}
