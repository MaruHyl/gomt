package gomt

import "errors"

type options struct {
	maxLevel int
	target   string
	json     bool
}

var defaultOptions = options{}

type Option func(opts *options) error

func buildOpts(options ...Option) (options, error) {
	opts := defaultOptions
	for _, opt := range options {
		err := opt(&opts)
		if err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// Set max level of tree
func WithMaxLevel(maxLevel int) Option {
	return func(opts *options) error {
		if maxLevel < 0 {
			return errors.New("max level must more than 0")
		}
		opts.maxLevel = maxLevel
		return nil
	}
}

// Draw mod graph only associated with target mod
func WithTarget(target string) Option {
	return func(opts *options) error {
		opts.target = target
		return nil
	}
}

// Draw mod graph use json
func WithJson(json bool) Option {
	return func(opts *options) error {
		opts.json = json
		return nil
	}
}
