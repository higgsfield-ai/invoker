package misc

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type SetupConfig struct {
	XAPIKey string `json:"xapikey"`
}

func (sc *SetupConfig) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.WithMessagef(err, "failed to open file %s", path)
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(sc); err != nil {
		return errors.WithMessagef(err, "failed to unmarshal json data from file %s", path)
	}

	return nil
}

func (sc *SetupConfig) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.WithMessagef(err, "failed to create file %s", path)
	}

	defer f.Close()

	if err = json.NewEncoder(f).Encode(sc); err != nil {
		return errors.WithMessagef(err, "failed to marshal json data to file %s", path)
	}

	return nil
}
