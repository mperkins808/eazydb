package eazydb

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Client struct {
	*sql.DB
	log *logrus.Logger
}

type ClientOptions struct {
	User       string
	Password   string
	Host       string
	Port       string
	Name       string
	Type       DB_TYPE
	Logger     *logrus.Logger
	EnableLogs bool
}

func NewClient(opts ...ClientOptions) (*Client, error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	logrus.New()
	dsn := fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=disable", opt.Host, opt.Port, opt.User, opt.Password, opt.Name)
	db, err := sql.Open(string(opt.Type), dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &Client{db, initLogger(opt.Logger, opt.EnableLogs)}, nil
}

func (c *Client) Test() {
	c.log.Info("hello world 123")
	c.log.Error("dasda")
}
