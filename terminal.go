package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/term"
)

var (
	stdinReader = bufio.NewReader(os.Stdin)

	termFD   = int(syscall.Stdin)
	rawMu    sync.Mutex
	rawState *term.State
)

func setupSignalHandling() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		rawMu.Lock()
		if rawState != nil {
			_ = term.Restore(termFD, rawState)
			rawState = nil
		}
		rawMu.Unlock()
		cancel()
		fmt.Fprintln(os.Stderr, "\ninterrupted")
		os.Exit(130)
	}()
	return ctx
}

func readPasswordHidden(prompt string) (string, error) {
	fmt.Print(prompt)

	oldState, err := term.GetState(termFD)
	if err != nil {
		return "", fmt.Errorf("stdin is not a terminal: %w", err)
	}

	rawMu.Lock()
	rawState = oldState
	rawMu.Unlock()

	b, err := term.ReadPassword(termFD)

	rawMu.Lock()
	rawState = nil
	rawMu.Unlock()

	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readLine() string {
	line, err := stdinReader.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimSpace(line)
}

func askYesNo(prompt string) bool {
	fmt.Print(prompt + " [y/N] ")
	answer := strings.ToLower(readLine())
	return answer == "y" || answer == "yes"
}

func isStdoutTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func isStdinTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
