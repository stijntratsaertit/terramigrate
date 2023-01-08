package generic

import (
	"fmt"
	"stijntratsaertit/terramigrate/config"
	"stijntratsaertit/terramigrate/database/adapter"
	"stijntratsaertit/terramigrate/database/postgres"
)

var supportedAdapters = map[string]func(c *config.DatabaseConnectionParams) (adapter.Adapter, error){
	"postgres": postgres.GetDatabase,
}

func GetDatabaseAdapter(adapter string) (adapter.Adapter, error) {
	dbCP, err := config.GetDatabaseConnectionParams()
	if err != nil {
		return nil, fmt.Errorf("could not get database connection params: %v", err)
	}

	if adapterFn, ok := supportedAdapters[adapter]; ok {
		return adapterFn(dbCP)
	}

	return nil, fmt.Errorf("unsupported adapter: %v", adapter)
}
