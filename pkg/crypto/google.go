package crypto

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/grpc/codes"
)

type GoogleProvider struct {
	Project  string
	Location string
	Keyring  string
	Key      string
	Options  []option.ClientOption
}

func (p GoogleProvider) keyName() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", p.Project, p.Location, p.Keyring, p.Key)
}

func googleErrorRetryable(err error) bool {
	var ae *apierror.APIError
	if _, ok := err.(net.Error); ok {
		return true
	} else if errors.As(err, &ae) {
		if ae.GRPCStatus() == nil {
			return true
		}
		code := ae.GRPCStatus().Code()
		return code == codes.Unavailable || code == codes.Aborted
	}
	return false
}

func (p GoogleProvider) Encrypt(plaintext string, retries uint, timeout time.Duration) ([]byte, error) {
	var result *kmspb.EncryptResponse
	f := func(ctx context.Context) error {
		client, err := kms.NewKeyManagementClient(ctx, p.Options...)
		if err != nil {
			return err
		}
		defer client.Close()
		result, err = client.Encrypt(ctx, &kmspb.EncryptRequest{
			Name:      p.keyName(),
			Plaintext: []byte(plaintext),
		})
		return err
	}
	err := retry(f, googleErrorRetryable, retries, timeout)
	if err != nil {
		return []byte{}, err
	} else {
		return result.Ciphertext, err
	}
}

func (p GoogleProvider) Decrypt(ciphertext []byte, retries uint, timeout time.Duration) (string, error) {
	var result *kmspb.DecryptResponse
	f := func(ctx context.Context) error {
		client, err := kms.NewKeyManagementClient(ctx, p.Options...)
		if err != nil {
			return err
		}
		defer client.Close()
		result, err = client.Decrypt(ctx, &kmspb.DecryptRequest{
			Name:       p.keyName(),
			Ciphertext: ciphertext,
		})
		return err
	}
	err := retry(f, googleErrorRetryable, retries, timeout)
	if err != nil {
		return "", err
	} else {
		return string(result.Plaintext), err
	}
}
