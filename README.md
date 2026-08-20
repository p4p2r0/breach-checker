# breach-checker

Password and email breach checker

## Why

Checking if a password or email address has been exposed in a breach helps decide whether it's still safe to use.

## How it works

This tool uses the [Have I Been Pwned](https://haveibeenpwned.com/) APIs to check passwords and email addresses for known breaches.

1. For password checks, the password is hashed locally using SHA-1.
2. The hash is split into a 5-character prefix and a 35-character suffix.
3. Only the prefix is sent to the HIBP Pwned Passwords API.
4. The API returns every suffix that matches that prefix.
5. The suffix is matched locally against the returned results.
6. For email checks, the email address is sent to the HIBP breach account API, which returns the known breaches associated with that address.

Because only 5 characters of a password's hash are sent to HIBP, the original password cannot be reconstructed from the request.

A password not found in the breach database is not confirmed to be strong. It only means that it does not appear in the dataset being checked.

Email breach checks require a HIBP API key.

## Usage

```
USAGE
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
```

## Installation

Requirements: Go 1.20+.

```bash
go install github.com/p4p2r0/breach-checker@latest
```

## License

This project is licensed under the [MIT License](LICENSE).
