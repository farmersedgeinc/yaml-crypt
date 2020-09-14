package crypto
import (
	"fmt"
)

type Provider interface {
	Encrypt(string) ([]byte, error)
	Decrypt([]byte) (string, error)
}

func getString(config map[string] interface{}, key string) (string, error) {
	value, ok := config[key]
	if !ok || value == "" { return "", fmt.Errorf("Required setting: .config.%s", key) }
	stringValue, ok := value.(string)
	if !ok || value == "" { return "", fmt.Errorf(".config.%s must be of type string", key) }
	return stringValue, nil
}

func NewProvider(name string, config map[string] interface{}) (Provider, error) {
	var provider Provider
	var err error
	switch name {
	case "noop":
		verbose, ok := config["verbose"]
		provider = NoopProvider{Verbose: ok && verbose.(bool)}
	case "google":
		project, err := getString(config, "project")
		if err != nil { err = err }
		location, err := getString(config, "location")
		if err != nil { err = err }
		keyring, err := getString(config, "keyring")
		if err != nil { err = err }
		key, err := getString(config, "key")
		if err != nil { err = err }
		provider = GoogleProvider{
			Project: project,
			Location: location,
			Keyring: keyring,
			Key: key,
		}
	default:
		err = fmt.Errorf("No provider named %s", name)
	}
	return provider, err
}

var BlankConfigs map[string]interface{} = map[string]interface{} {
	"noop": map[string]interface{} {
		"verbose": false,
	},
	"google": map[string]interface{} {
		"project": "",
		"location": "global",
		"keyring": "",
		"key": "",
	},
}
