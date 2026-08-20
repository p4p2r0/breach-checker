package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func colorize(useColor bool, code, s string) string {
	if !useColor {
		return s
	}
	return code + s + colorReset
}

func printHuman(pwResults []PasswordResult, emResults []EmailResult, useColor bool) {
	if len(pwResults) > 0 {
		fmt.Println(colorize(useColor, colorBold, "Passwords"))
		for _, r := range pwResults {
			printPasswordResult(r, useColor)
		}
		if len(emResults) > 0 {
			fmt.Println()
		}
	}

	if len(emResults) > 0 {
		fmt.Println(colorize(useColor, colorBold, "Emails"))
		for _, r := range emResults {
			printEmailResult(r, useColor)
		}
	}
}

func printPasswordResult(r PasswordResult, useColor bool) {
	prefix := fmt.Sprintf("  %-14s %s", r.Label, colorize(useColor, colorDim, r.Hash))
	switch {
	case r.Err != "":
		fmt.Printf("%s  %s — %s\n", prefix, colorize(useColor, colorYellow, "ERROR"), r.Err)
	case r.Pwned:
		fmt.Printf("%s  %s — seen %d time(s) in known breaches\n", prefix, colorize(useColor, colorRed, "PWNED"), r.Count)
	default:
		fmt.Printf("%s  %s\n", prefix, colorize(useColor, colorGreen, "not found"))
	}
}

func printEmailResult(r EmailResult, useColor bool) {
	prefix := fmt.Sprintf("  %s", r.Email)
	switch {
	case r.Err != "":
		fmt.Printf("%s  %s — %s\n", prefix, colorize(useColor, colorYellow, "ERROR"), r.Err)
	case r.Count > 0:
		names := make([]string, len(r.Breaches))
		for i, b := range r.Breaches {
			title := b.Title
			if title == "" {
				title = b.Name
			}
			if b.BreachDate != "" {
				names[i] = fmt.Sprintf("%s (%s)", title, b.BreachDate)
			} else {
				names[i] = title
			}
		}
		fmt.Printf("%s  %s — found in %d breach(es): %s\n", prefix, colorize(useColor, colorRed, "PWNED"), r.Count, strings.Join(names, ", "))
	default:
		fmt.Printf("%s  %s\n", prefix, colorize(useColor, colorGreen, "not found"))
	}
}

type jsonSummary struct {
	Checked int `json:"checked"`
	Pwned   int `json:"pwned"`
	Errors  int `json:"errors"`
}

type jsonReport struct {
	Passwords []PasswordResult `json:"passwords,omitempty"`
	Emails    []EmailResult    `json:"emails,omitempty"`
	Summary   jsonSummary      `json:"summary"`
}

func printJSON(pwResults []PasswordResult, emResults []EmailResult) {
	report := jsonReport{Passwords: pwResults, Emails: emResults}
	for _, r := range pwResults {
		report.Summary.Checked++
		if r.Err != "" {
			report.Summary.Errors++
		} else if r.Pwned {
			report.Summary.Pwned++
		}
	}
	for _, r := range emResults {
		report.Summary.Checked++
		if r.Err != "" {
			report.Summary.Errors++
		} else if r.Count > 0 {
			report.Summary.Pwned++
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(report)
}

func computeExitCode(pwResults []PasswordResult, emResults []EmailResult) int {
	hasErr := false
	hasPwned := false
	for _, r := range pwResults {
		if r.Err != "" {
			hasErr = true
		}
		if r.Pwned {
			hasPwned = true
		}
	}
	for _, r := range emResults {
		if r.Err != "" {
			hasErr = true
		}
		if r.Count > 0 {
			hasPwned = true
		}
	}
	switch {
	case hasErr:
		return 2
	case hasPwned:
		return 1
	default:
		return 0
	}
}
