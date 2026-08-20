package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

type options struct {
	passwords    multiFlag
	passwordFile string
	stdin        bool

	emails    multiFlag
	emailFile string

	apiKey  string
	json    bool
	noColor bool
	help    bool
	version bool

	positional []string
}

func parseFlags(args []string) *options {
	fs := flag.NewFlagSet("breach-checker", flag.ExitOnError)
	fs.Usage = func() { printUsage(fs.Output()) }

	opts := &options{}

	fs.Var(&opts.passwords, "p", "")
	fs.Var(&opts.passwords, "password", "")
	fs.StringVar(&opts.passwordFile, "password-file", "", "")
	fs.BoolVar(&opts.stdin, "stdin", false, "")

	fs.Var(&opts.emails, "e", "")
	fs.Var(&opts.emails, "email", "")
	fs.StringVar(&opts.emailFile, "email-file", "", "")

	fs.StringVar(&opts.apiKey, "k", "", "")
	fs.StringVar(&opts.apiKey, "api-key", "", "")

	fs.BoolVar(&opts.json, "json", false, "")
	fs.BoolVar(&opts.noColor, "no-color", false, "")
	fs.BoolVar(&opts.help, "h", false, "")
	fs.BoolVar(&opts.help, "help", false, "")
	fs.BoolVar(&opts.version, "v", false, "")
	fs.BoolVar(&opts.version, "version", false, "")

	_ = fs.Parse(args)

	if opts.help {
		printUsage(os.Stdout)
		os.Exit(0)
	}
	if opts.version {
		fmt.Printf("breach-checker %s\n", version)
		os.Exit(0)
	}

	opts.positional = fs.Args()
	return opts
}

func printUsage(w interface{ Write([]byte) (int, error) }) {
	fmt.Fprint(w, `USAGE
  breach-checker [flags] [password ...]

FLAGS
  -p, --password STRING       Password to check (repeatable)
      --password-file PATH    Read passwords from a file, one per line
      --stdin                 Read passwords from stdin, one per line

  -e, --email STRING          Email address to check (repeatable)
      --email-file PATH       Read email addresses from a file, one per line

  -k, --api-key STRING        HIBP API key for email checks (env HIBP_API_KEY)

      --json                  Output machine-readable JSON
      --no-color              Disable colored output
  -h, --help                  Show this help message
  -v, --version                Show version information

  Bare positional arguments are treated as passwords. If nothing is
  supplied at all and stdin is a terminal, you'll be prompted interactively
  with input hidden (like sudo).

EXAMPLES
  breach-checker
  breach-checker hunter2 correcthorsebatterystaple
  breach-checker -p hunter2 -p correcthorsebatterystaple
  breach-checker --password-file passwords.txt --json
  cat passwords.txt | breach-checker --stdin
  breach-checker -e alice@example.com -e bob@example.com --api-key XXXX

EXIT CODES
  0  no breaches found
  1  at least one password or email was found in a breach
  2  an error occurred (network, invalid arguments, missing API key, ...)
`)
}

func readLinesFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	defer f.Close()
	return scanLines(f), nil
}

func readLinesStdin() []string {
	return scanLines(os.Stdin)
}

func scanLines(r interface{ Read([]byte) (int, error) }) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
