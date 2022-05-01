package git

import (
	"context"
	"database/sql"
	"errors"
)

type Config struct {
	baseBranch    sql.NullString
	featureBranch sql.NullString
}

type Option func(*Config)

func BaseBranch(baseBranch string) Option {
	return func(c *Config) {
		c.baseBranch = sql.NullString{String: baseBranch, Valid: true}
	}
}

func FeatureBranch(featureBranch string) Option {
	return func(c *Config) {
		c.featureBranch = sql.NullString{String: featureBranch, Valid: true}
	}
}

func (c *Config) Add(opts ...Option) *Config {
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Config) Validate(ctx context.Context) (err error) {
	switch {
	case c.baseBranch.Valid && c.featureBranch.Valid:
		c.baseBranch.String, c.featureBranch.String, err = FindBaseAndFeatureBranches(
			ctx, c.baseBranch.String, c.featureBranch.String)
		if err != nil {
			return
		}
	case c.featureBranch.Valid:
		c.featureBranch.String, err = FindFeatureBranch(ctx, c.featureBranch.String)
		if err != nil {
			return
		}
	}

	return nil
}

func (c *Config) FeatureBranch() (string, error) {
	if !c.featureBranch.Valid {
		return "", errors.New("feature branch is undefined")
	}

	return c.featureBranch.String, nil
}

func (c *Config) Branches() (baseBranch, featureBranch string, err error) {
	if !c.baseBranch.Valid {
		return "", "", errors.New("base branch is undefined")
	}

	if !c.featureBranch.Valid {
		return "", "", errors.New("feature branch is undefined")
	}

	return c.baseBranch.String, c.featureBranch.String, nil
}
