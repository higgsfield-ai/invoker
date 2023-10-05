package internal

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type MetadataManager struct {
	path     string
	metadata map[string]interface{}
}

func NewMetadataManager(path string) (*MetadataManager, error) {
	// check if exists
	_, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.WithMessage(err, "failed to stat metadata file")
		} else {
			return &MetadataManager{
				path:     path,
				metadata: nil,
			}, nil
		}
	}

	s := make(map[string]interface{})
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open metadata file")
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&s); err != nil {
		return nil, errors.WithMessage(err, "failed to decode metadata file")
	}
	return &MetadataManager{
		path:     path,
		metadata: s,
	}, nil
}

func (m *MetadataManager) KillExperiments(experimentName *string) ([]string, error) {
	if m.metadata == nil {
		return nil, nil
	}

	if m.metadata["experiments"] == nil {
		return nil, nil
	}

	killed := make([]string, 0)
	// if experimentName is nil, kill all experiments
	// else kill only the experiment with the given name
	if experimentName != nil || *experimentName != "" {
		experiment := m.metadata["experiments"].(map[string]interface{})[*experimentName]
		if experiment == nil {
			return nil, nil
		}

		casted := (experiment).(map[string]interface{})
		kill(&casted)
		killed = append(killed, *experimentName)
	} else {
		for experimentName, experiment := range m.metadata["experiments"].(map[string]interface{}) {
			casted := (experiment).(map[string]interface{})
			kill(&casted)
			killed = append(killed, experimentName)
		}
	}

	// save metadata
	res, err := json.MarshalIndent(m.metadata, "", "  ")
	if err != nil {
		return killed, errors.WithMessage(err, "failed to marshal metadata")
	}

	if err := os.WriteFile(m.path, res, 0644); err != nil {
		return killed, errors.WithMessage(err, "failed to write metadata file")
	}

	return killed, nil
}

func kill(experiment *map[string]interface{}) {
	if (*experiment)["runs"] == nil {
		return
	}
	runs := (*experiment)["runs"].(map[string]interface{})
	for _, run := range runs {
		run := run.(map[string]interface{})
		state := run["state"].(string)
		if state == "ABORTED" || state == "COMPLETED" {
			continue
		}

		run["state"] = "ABORTED"
	}
}
