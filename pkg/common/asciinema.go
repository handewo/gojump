package common

import (
	"encoding/json"
	"io"
	"time"
)

const (
	version      = 2
	defaultShell = "/bin/bash"
	defaultTerm  = "xterm"
)

var (
	newLine = []byte{'\n'}
)

func NewWriter(w io.Writer, opts ...AsciiOption) *AsciiWriter {
	conf := AsciiConfig{
		Width:    80,
		Height:   40,
		EnvShell: defaultShell,
		EnvTerm:  defaultTerm,
	}
	for _, setter := range opts {
		setter(&conf)
	}
	return &AsciiWriter{
		AsciiConfig:   conf,
		TimestampNano: conf.Timestamp.UnixNano(),
		writer:        w,
	}
}

type AsciiWriter struct {
	AsciiConfig
	TimestampNano int64
	writer        io.Writer
}

func (w *AsciiWriter) WriteHeader() error {
	header := Header{
		Version:   version,
		Width:     w.Width,
		Height:    w.Height,
		Timestamp: w.Timestamp.Unix(),
		Title:     w.Title,
		Env: Env{
			Shell: w.EnvShell,
			Term:  w.EnvTerm,
		},
	}
	raw, err := json.Marshal(header)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(raw)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(newLine)
	return err
}

func (w *AsciiWriter) WriteRow(p []byte) error {
	now := time.Now().UnixNano()
	ts := float64(now-w.TimestampNano) / 1000 / 1000 / 1000
	return w.WriteStdout(ts, p)
}

func (w *AsciiWriter) WriteStdout(ts float64, data []byte) error {
	row := []interface{}{ts, "o", string(data)}
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(raw)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(newLine)
	return err
}

type Header struct {
	Version   int    `json:"version"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title"`
	Env       Env    `json:"env"`
}

type Env struct {
	Shell string `json:"SHELL"`
	Term  string `json:"TERM"`
}

type AsciiConfig struct {
	Title     string
	EnvShell  string
	EnvTerm   string
	Width     int
	Height    int
	Timestamp time.Time
}

type AsciiOption func(options *AsciiConfig)

func WithWidth(width int) AsciiOption {
	return func(options *AsciiConfig) {
		options.Width = width
	}
}

func WithHeight(height int) AsciiOption {
	return func(options *AsciiConfig) {
		options.Height = height
	}
}

func WithTimestamp(timestamp time.Time) AsciiOption {
	return func(options *AsciiConfig) {
		options.Timestamp = timestamp
	}
}

func WithTitle(title string) AsciiOption {
	return func(options *AsciiConfig) {
		options.Title = title
	}
}

func WithEnvShell(shell string) AsciiOption {
	return func(options *AsciiConfig) {
		options.EnvShell = shell
	}
}

func WithEnvTerm(term string) AsciiOption {
	return func(options *AsciiConfig) {
		options.EnvTerm = term
	}
}
