package collector

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type LogParams struct {
	ExperimentName string `json:"experiment_name"`
	ProjectName    string `json:"project_name"`
	RunName        string `json:"run_name"`
	ContainerID    string `json:"container_id"`
	NodeRank       string `json:"node_rank"`
}

type LogToCollect struct {
	LogParams
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
	Line      string `json:"line"`
}

type StopOrErr struct {
	LogParams
	StopOrErr string `json:"stop_or_err"`
}

type CtxKey string

const (
	Txn CtxKey = "tx"
)

func decodeLogsToCollect(r *http.Request) ([]LogToCollect, error) {
	var logs []LogToCollect
	if err := json.NewDecoder(r.Body).Decode(&logs); err != nil {
		return nil, errors.WithMessage(err, "failed to decode logs to collect")
	}

	return logs, nil
}

func decodeStopOrErr(r *http.Request) (*StopOrErr, error) {
	var soe StopOrErr
	if err := json.NewDecoder(r.Body).Decode(&soe); err != nil {
		return nil, errors.WithMessage(err, "failed to decode stop or err")
	}

	return &soe, nil
}
