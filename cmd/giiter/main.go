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
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := exec.CommandContext(ctx, "git", "log", `--pretty=format:%H`)

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
