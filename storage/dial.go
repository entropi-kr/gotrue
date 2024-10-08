package storage

import (
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	"github.com/entropi-kr/gotrue/conf"
	"github.com/entropi-kr/gotrue/storage/namespace"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/pop/v5/columns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/url"
)

// Connection is the interface a storage provider must implement.
type Connection struct {
	*pop.Connection
}

// Dial will connect to that storage engine
func Dial(config *conf.GlobalConfiguration) (*Connection, error) {
	if config.DB.Driver == "" && config.DB.URL != "" {
		u, err := url.Parse(config.DB.URL)
		if err != nil {
			return nil, errors.Wrap(err, "parsing db connection url")
		}
		config.DB.Driver = u.Scheme
	}

	namespace.SetNamespace(config.DB.Namespace)

	driver := ""
	if config.DB.Driver != "postgres" {
		logrus.Warn("DEPRECATION NOTICE: only PostgreSQL is supported by entropi GoTrue, will be removed soon")
	} else {
		// pop v5 uses pgx as the default PostgreSQL driver
		driver = "pgx"
	}

	options := make(map[string]string)

	db, err := pop.NewConnection(&pop.ConnectionDetails{
		Dialect: config.DB.Driver,
		Driver:  driver,
		URL:     config.DB.URL,
		Options: options,
	})
	if err != nil {
		return nil, errors.Wrap(err, "opening database connection")
	}
	if err := db.Open(); err != nil {
		return nil, errors.Wrap(err, "checking database connection")
	}

	return &Connection{db}, nil
}

func (c *Connection) Transaction(fn func(*Connection) error) error {
	if c.TX == nil {
		return c.Connection.Transaction(func(tx *pop.Connection) error {
			return fn(&Connection{tx})
		})
	}
	return fn(c)
}

func getExcludedColumns(model interface{}, includeColumns ...string) ([]string, error) {
	sm := &pop.Model{Value: model}

	// get all columns and remove included to get excluded set
	cols := columns.ForStructWithAlias(model, sm.TableName(), sm.As, sm.IDField())
	for _, f := range includeColumns {
		if _, ok := cols.Cols[f]; !ok {
			return nil, errors.Errorf("Invalid column name %s", f)
		}
		cols.Remove(f)
	}

	xcols := make([]string, len(cols.Cols))
	for n := range cols.Cols {
		xcols = append(xcols, n)
	}
	return xcols, nil
}
