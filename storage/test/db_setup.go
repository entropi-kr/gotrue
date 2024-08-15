package test

import (
	"gitlab.com/entropi-tech/gotrue/conf"
	"gitlab.com/entropi-tech/gotrue/storage"
)

func SetupDBConnection(globalConfig *conf.GlobalConfiguration) (*storage.Connection, error) {
	return storage.Dial(globalConfig)
}
