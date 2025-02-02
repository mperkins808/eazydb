package eazydb

import (
	"fmt"
	"os"
)

func validateOptions(opt *ClientOptions) error {
	if opt.User == "" {
		return fmt.Errorf("User is not set, either pass as a client option or set DB_USER")
	}
	if opt.Password == "" {
		return fmt.Errorf("Password is not set, either pass as a client option or set DB_PASS")
	}
	if opt.Host == "" {
		return fmt.Errorf("Host is not set, either pass as a client option or set DB_HOST")
	}
	if opt.Port == "" {
		return fmt.Errorf("Port is not set, either pass as a client option or set DB_PORT")
	}
	if opt.Name == "" {
		return fmt.Errorf("Database name is not set, either pass as a client option or set DB_NAME")
	}
	return nil
}

func ifNoEnv(env string, fallback string) string {
	if os.Getenv(env) == "" {
		return fallback
	}
	return os.Getenv(env)
}

func parseOptions(opts ...ClientOptions) (*ClientOptions, error) {
	var opt ClientOptions
	if len(opts) > 0 {
		opt = opts[0]
		opt.User = ifNoEnv("DB_USER", opt.User)
		opt.Password = ifNoEnv("DB_PASS", opt.Password)
		opt.Host = ifNoEnv("DB_HOST", opt.Host)
		opt.Port = ifNoEnv("DB_PORT", opt.Port)
		opt.Name = ifNoEnv("DB_NAME", opt.Name)
		opt.Type = DB_TYPE(ifNoEnv("DB_TYPE", string(opt.Type)))
	} else {
		opt = ClientOptions{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
			Type:     DB_TYPE(os.Getenv("DB_TYPE")),
		}
	}

	if err := validateOptions(&opt); err != nil {
		return nil, err
	}

	return &opt, nil
}
