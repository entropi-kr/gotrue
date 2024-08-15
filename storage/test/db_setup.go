package test

import (
	"github.com/entropi-kr/gotrue/conf"
	"github.com/entropi-kr/gotrue/storage"
)

func SetupDBConnection(globalConfig *conf.GlobalConfiguration) (*storage.Connection, error) {
	return storage.Dial(globalConfig)
}
