package cassandra

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
)

func InitCluster(connection string) (*gocql.ClusterConfig, error) {
	connectionURL, err := url.Parse(connection)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect Database connection string format shuld be cassandra://[user:password@]ip1[:port]/keyspace[?dc=dc_name] but have %v", connection)
	}

	cluster := gocql.NewCluster(connectionURL.Host)
	cluster.Keyspace = strings.TrimLeft(connectionURL.Path, "/")
	if len(cluster.Keyspace) == 0 {
		return nil, errors.New(fmt.Sprintf("Connection string must content a keyspace"))
	}
	if connectionURL.User != nil {
		pwd, exist := connectionURL.User.Password()
		if !exist {
			return nil, errors.New(fmt.Sprintf("Connection string must content a password for user"))
		}
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: connectionURL.User.Username(),
			Password: pwd,
		}
	}
	if connectionURL.Query().Get("dc") != "" {
		cluster.PoolConfig = gocql.PoolConfig{
			HostSelectionPolicy: gocql.DCAwareRoundRobinPolicy(connectionURL.Query().Get("dc")),
		}
	}
	if connectionURL.Query().Get("init_host_lookup") != "" {
		cluster.DisableInitialHostLookup = connectionURL.Query().Get("init_host_lookup") == "false"
	}
	return cluster, nil
}
