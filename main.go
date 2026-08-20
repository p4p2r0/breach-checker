package main

import (
	"fmt"
	"os"
)

func main() {
	loadDotEnv(".env")

	opts := parseFlags(os.Args[1:])
	ctx := setupSignalHandling()

	apiKey := resolveAPIKey(opts.apiKey)

	passwords, err := gatherPasswords(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	emails := []string(opts.emails)
	if opts.emailFile != "" {
		more, err := readLinesFile(opts.emailFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		emails = append(emails, more...)
	}

	if len(passwords) == 0 && len(emails) == 0 {
		if isStdinTerminal() {
			passwords, emails = runInteractive(apiKey)
		} else {
			fmt.Fprintln(os.Stderr, "error: nothing to check — provide -p/-e, --stdin, --password-file/--email-file, or run interactively from a terminal")
			os.Exit(2)
		}
	}

	if len(passwords) == 0 && len(emails) == 0 {
		fmt.Fprintln(os.Stderr, "error: nothing to check")
		os.Exit(2)
	}

	pwResults := checkPasswords(ctx, passwords)
	var emResults []EmailResult
	if len(emails) > 0 {
		emResults = checkEmails(ctx, emails, apiKey)
	}

	if opts.json {
		printJSON(pwResults, emResults)
	} else {
		useColor := !opts.noColor && isStdoutTerminal()
		printHuman(pwResults, emResults, useColor)
	}

	os.Exit(computeExitCode(pwResults, emResults))
}

func gatherPasswords(opts *options) ([]string, error) {
	var passwords []string
	passwords = append(passwords, opts.positional...)
	passwords = append(passwords, opts.passwords...)

	if opts.passwordFile != "" {
		fromFile, err := readLinesFile(opts.passwordFile)
		if err != nil {
			return nil, err
		}
		passwords = append(passwords, fromFile...)
	}

	if opts.stdin {
		passwords = append(passwords, readLinesStdin()...)
	}

	return passwords, nil
}

func runInteractive(apiKey string) (passwords, emails []string) {
	fmt.Println("What would you like to check?")
	fmt.Println("  [1] Password")
	fmt.Println("  [2] Email address")
	fmt.Println("  [3] Both")

	var choice string
	for {
		fmt.Print("> ")
		choice = readLine()
		if choice == "1" || choice == "2" || choice == "3" {
			break
		}
		fmt.Println("Please enter 1, 2, or 3.")
	}

	if choice == "1" || choice == "3" {
		passwords = collectPasswords()
	}
	if choice == "2" || choice == "3" {
		emails = collectEmails(apiKey)
	}

	return passwords, emails
}

func collectPasswords() []string {
	var passwords []string
	for {
		pw, err := readPasswordHidden("Enter password: ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading password:", err)
			os.Exit(2)
		}
		if pw == "" {
			fmt.Fprintln(os.Stderr, "error: password cannot be empty")
			os.Exit(2)
		}
		passwords = append(passwords, pw)
		if !askYesNo("Check another password?") {
			break
		}
	}
	return passwords
}

func collectEmails(apiKey string) []string {
	if apiKey == "" {
		fmt.Println("No HIBP API key configured (--api-key, HIBP_API_KEY, or .env) — cannot check emails.")
		fmt.Println("Get a free-tier key at https://haveibeenpwned.com/API/Key")
		return nil
	}
	var emails []string
	for {
		fmt.Print("Enter email address: ")
		email := readLine()
		if email != "" {
			emails = append(emails, email)
		}
		if !askYesNo("Check another email?") {
			break
		}
	}
	return emails
}
