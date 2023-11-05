package slack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"go.alexhamlin.co/randomizer/internal/awsconfig"
)

const DefaultAWSParameterTTL = 2 * time.Minute

// TokenProvider provides the value of the slash command verification token
// generated by Slack.
type TokenProvider func(ctx context.Context) (string, error)

// TokenProviderFromEnv returns a TokenProvider based on available environment
// variables.
//
// If SLACK_TOKEN is set, it will return a static token provider.
//
// If SLACK_TOKEN_SSM_NAME is set, it will return an AWS SSM token provider,
// with the TTL optionally set by SLACK_TOKEN_SSM_TTL.
//
// Otherwise, it will return an error.
func TokenProviderFromEnv() (TokenProvider, error) {
	if token, ok := os.LookupEnv("SLACK_TOKEN"); ok {
		return StaticToken(token), nil
	}

	if ssmName, ok := os.LookupEnv("SLACK_TOKEN_SSM_NAME"); ok {
		ttl, err := ssmTTLFromEnv()
		if err != nil {
			return nil, err
		}
		return AWSParameter(ssmName, ttl), nil
	}

	return nil, errors.New("missing SLACK_TOKEN or SLACK_TOKEN_SSM_NAME in environment")
}

func ssmTTLFromEnv() (time.Duration, error) {
	ttlEnv, ok := os.LookupEnv("SLACK_TOKEN_SSM_TTL")
	if !ok {
		return DefaultAWSParameterTTL, nil
	}

	ttl, err := time.ParseDuration(ttlEnv)
	if err != nil {
		return 0, fmt.Errorf("SLACK_TOKEN_SSM_TTL is not a valid Go duration: %w", err)
	}

	return ttl, nil
}

// StaticToken uses token as the expected value of the verification token.
func StaticToken(token string) TokenProvider {
	return func(_ context.Context) (string, error) {
		return token, nil
	}
}

// AWSParameter retrieves the expected value of the verification token from the
// AWS SSM Parameter Store, decrypting it if necessary, and caches the retrieved
// token value for the provided TTL.
func AWSParameter(name string, ttl time.Duration) TokenProvider {
	var (
		mu     ctxLock
		token  string
		expiry time.Time
	)

	return func(ctx context.Context) (string, error) {
		if !mu.LockWithContext(ctx) {
			return "", ctx.Err()
		}
		defer mu.Unlock()

		if time.Now().Before(expiry) {
			return token, nil
		}

		cfg, err := awsconfig.New(ctx)
		if err != nil {
			return "", err
		}

		output, err := ssm.NewFromConfig(cfg).GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(name),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return "", fmt.Errorf("loading Slack token parameter: %w", err)
		}

		token = *output.Parameter.Value
		expiry = time.Now().Add(ttl)
		return token, nil
	}
}

// ctxLock is a mutual exclusion lock that allows clients to cancel a pending
// lock operation with a context. The zero value is an unlocked lock.
type ctxLock struct {
	init sync.Once
	ch   chan struct{} // buffered, size 1
}

// LockWithContext attempts to lock l. If the lock is already in use, the
// calling goroutine blocks until l is available or ctx is canceled. The return
// value indicates whether the lock was actually acquired; if false, ctx was
// canceled before the lock was acquired, and the caller must not unlock l or
// violate any invariant that l protects.
func (l *ctxLock) LockWithContext(ctx context.Context) bool {
	l.ensureInit()
	select {
	case <-l.ch:
		return true
	case <-ctx.Done():
		return false
	}
}

// Unlock unlocks l.
func (l *ctxLock) Unlock() {
	l.ensureInit()
	select {
	case l.ch <- struct{}{}:
		return
	default:
		panic("unlock of unlocked ctxLock")
	}
}

func (l *ctxLock) ensureInit() {
	l.init.Do(func() {
		l.ch = make(chan struct{}, 1)
		l.ch <- struct{}{}
	})
}
