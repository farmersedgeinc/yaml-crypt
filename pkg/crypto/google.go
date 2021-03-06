package crypto

import (
	kms "cloud.google.com/go/kms/apiv1"
	"context"
	"fmt"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
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

func (p GoogleProvider) Encrypt(plaintext string) ([]byte, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx, p.Options...)
	if err != nil {
		return []byte{}, err
	}
	defer client.Close()
	result, err := client.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:      p.keyName(),
		Plaintext: []byte(plaintext),
	})
	if err != nil {
		return []byte{}, err
	} else {
		return result.Ciphertext, err
	}
}

func (p GoogleProvider) Decrypt(ciphertext []byte) (string, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx, p.Options...)
	if err != nil {
		return "", err
	}
	defer client.Close()
	result, err := client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       p.keyName(),
		Ciphertext: ciphertext,
	})
	if err != nil {
		return "", err
	} else {
		return string(result.Plaintext), err
	}
}
