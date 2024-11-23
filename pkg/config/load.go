package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var glbEnvs map[string]string

func init() {
	glbEnvs = make(map[string]string)
	envs := os.Environ()
	for _, env := range envs {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		glbEnvs[pair[0]] = pair[1]
	}
}

// LoadConfigure loads configuration from bytes and unmarshal into c.
// Now it supports json, yaml and toml format.
func LoadConfigure(b []byte, c any, strict bool) error {
	var tomlObj interface{}
	// Try to unmarshal as TOML first; swallow errors from that (assume it's not valid TOML).
	if err := toml.Unmarshal(b, &tomlObj); err == nil {
		b, err = json.Marshal(&tomlObj)
		if err != nil {
			return err
		}
	}
	// If the buffer smells like JSON (first non-whitespace character is '{'), unmarshal as JSON directly.
	if yaml.IsJSONBuffer(b) {
		decoder := json.NewDecoder(bytes.NewBuffer(b))
		if strict {
			decoder.DisallowUnknownFields()
		}
		return decoder.Decode(c)
	}
	// It wasn't JSON. Unmarshal as YAML.
	if strict {
		return yaml.UnmarshalStrict(b, c)
	}
	return yaml.Unmarshal(b, c)
}
