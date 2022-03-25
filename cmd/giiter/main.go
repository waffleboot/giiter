package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	m := newManager()

	app := &cli.App{
		Name: "giiter",
		Commands: []*cli.Command{
			{
				Name: "git",
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Action: m.gitList,
					},
				},
			},
		},
	}

	return app.RunContext(ctx, os.Args)
}

type manager struct{}

func newManager() *manager {
	return &manager{}
}

func (m *manager) gitList(ctx *cli.Context) error {
	cmd := exec.CommandContext(ctx.Context, "git", "log", `--pretty=format:%h`)

	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	s := bufio.NewScanner(bytes.NewReader(stdout))

	for s.Scan() {
		fmt.Println(s.Text())
	}

	if err := s.Err(); err != nil {
		return err
	}

	return nil
}
