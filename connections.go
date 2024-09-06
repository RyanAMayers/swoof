package main

import (
	"net/url"
	"os"

	mysql "github.com/StirlingMarketingGroup/cool-mysql"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type connection struct {
	User   string            `yaml:"user"`
	Pass   string            `yaml:"pass"`
	Host   string            `yaml:"host"`
	Schema string            `yaml:"schema"`
	Params map[string]string `yaml:"params"`

	SourceOnly bool `yaml:"source_only"`
	DestOnly   bool `yaml:"dest_only"`
}

// GetConnections returns our connection map that's
// parsed from the user's config dir,
// makes calls to swoof much shorter and much easier
// and even a little safer potentially
func GetConnections(file string) (connections map[string]connection, err error) {
	y, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(y, &connections)
	if err != nil {
		return nil, err
	}

	return
}

// ConnectionToDSN converts our own connection structs to the
// official mysql's own connection struct, formatted for our use case
func ConnectionToDSN(c connection) string {
	d := mysqldriver.NewConfig()
	d.User = c.User
	d.Passwd = c.Pass
	d.Net = "tcp"
	d.Addr = c.Host
	d.DBName = c.Schema

	dsn := d.FormatDSN()
	i := 0
	for k, v := range c.Params {
		if i == 0 {
			dsn += "?"
		} else {
			dsn += "&"
		}
		dsn += url.QueryEscape(k) + "=" + url.QueryEscape(v)
		i++
	}

	return dsn
}

// CheckIfInSource is a wrapper function that checks if the
// the table for a given connection exists and panics if
// if there is no table in that connection
func CheckIfInSource(s *mysql.Database, t string) {
	if ok, err := s.Exists("show tables like'"+t+"'", 0); err != nil {
		panic(err)
	} else if !ok {
		panic(errors.Errorf("table %q does not exist on the source connection", t))
	}
}

func TestConnection(c connection) {
	dsn := ConnectionToDSN(c)
	s, err := mysql.NewFromDSN(dsn, dsn)
	if err != nil {
		panic(err)
	}
	if err := s.Test(); err != nil {
		panic(err)
	}
}
