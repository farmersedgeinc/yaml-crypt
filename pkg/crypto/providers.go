package crypto

import (
	"context"
	"fmt"
	"os"
	"time"
)

type Provider interface {
	Encrypt(string, uint, time.Duration) ([]byte, error)
	Decrypt([]byte, uint, time.Duration) (string, error)
}

func getString(config map[string]interface{}, key string) (string, error) {
	value, ok := config[key]
	if !ok || value == "" {
		return "", fmt.Errorf("Required setting: .config.%s", key)
	}
	stringValue, ok := value.(string)
	if !ok || value == "" {
		return "", fmt.Errorf(".config.%s must be of type string", key)
	}
	return stringValue, nil
}

func NewProvider(name string, config map[string]interface{}) (Provider, error) {
	var provider Provider
	var err error
	switch name {
	case "noop":
		verbose, ok := config["verbose"]
		provider = NoopProvider{Verbose: ok && verbose.(bool)}
	case "google":
		project, err := getString(config, "project")
		if err != nil {
			return nil, err
		}
		location, err := getString(config, "location")
		if err != nil {
			return nil, err
		}
		keyring, err := getString(config, "keyring")
		if err != nil {
			return nil, err
		}
		key, err := getString(config, "key")
		if err != nil {
			return nil, err
		}
		provider = GoogleProvider{
			Project:  project,
			Location: location,
			Keyring:  keyring,
			Key:      key,
		}
	default:
		return nil, fmt.Errorf("No provider named %s", name)
	}
	return provider, err
}

var BlankConfigs map[string]interface{} = map[string]interface{}{
	"noop": map[string]interface{}{
		"verbose": false,
	},
	"google": map[string]interface{}{
		"project":  "",
		"location": "global",
		"keyring":  "",
		"key":      "",
	},
}

type Operation func(context.Context) error

func retry(operation Operation, isRetryable func(error) bool, retries uint, timeout time.Duration) error {
	var err error
	for i := uint(0); i < retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = operation(ctx)
		if err != nil {
			if !isRetryable(err) {
				return err
			}
		} else if err = ctx.Err(); err == nil {
			return err
		}
		if i < retries-1 {
			fmt.Fprintf(os.Stderr, "%s", fmt.Errorf("WARNING: crypto operation failed (try %d of %d) with error %w\n", i+1, retries, err))
			time.Sleep(time.Duration(200 * time.Millisecond * 1 << i))
		}
	}
	return err
}
