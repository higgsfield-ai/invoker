package collector

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ml-doom/invoker/internal/misc"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type ServeArgs struct {
	ProjectName string `validate:"required,varname"`
	Port        int    `validate:"required,min=1"`
}

var (
	ErrPortInUse          = errors.New("port is already in use")
	ErrPortInUseByInvoker = errors.New("port is already in use by invoker")
)

func Initialize(
	args ServeArgs,
) (*http.Server, error) {
	if err := misc.Validator().Struct(args); err != nil {
		return nil, errors.WithMessage(err, "invalid args")
	}

	cfgPath, hfDir, err := misc.SetupCfgHFDir(args.ProjectName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to setup project folder")
	}

	var cfg misc.SetupConfig

	// check if cfg file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// cfg file doesn't exist, so create it
		cfg.XAPIKey = misc.RandXAPIKey()
		if err = cfg.Save(cfgPath); err != nil {
			return nil, errors.WithMessage(err, "failed to save cfg file")
		}
	} else {
		if err := cfg.Load(cfgPath); err != nil {
			return nil, errors.WithMessage(err, "failed to load cfg file")
		}
	}

	sqliteDB, err := sql.Open("sqlite3", misc.DSN(hfDir+"/hf_logs.db"))
	if err != nil {
		return nil, errors.WithMessagef(err, "can't init db from given path %s", hfDir)
	}

	ctx := context.Background()

	db := NewDB(sqliteDB)

	if err = db.Warmup(ctx); err != nil {
		return nil, errors.WithMessage(err, "failed to warmup db")
	}

	r := NewRouter(cfg.XAPIKey, db)
	h2s := &http2.Server{}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", args.Port),
		Handler: h2c.NewHandler(r, h2s),
	}

	tab, err := ProcOnPort(uint16(args.Port))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to find process on port")
	}

	if tab != nil {
		if tab.Process.Name != "invoker" {
			return nil, errors.WithMessagef(ErrPortInUse, "port %d use by %-16s", args.Port, tab.Process.String())
		} else {
			return nil, ErrPortInUseByInvoker
		}
	}

	return server, nil
}
