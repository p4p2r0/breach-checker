package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	hibpPasswordsAPI          = "https://api.pwnedpasswords.com/range/"
	maxConcurrentPasswordReqs = 8
)

type PasswordResult struct {
	Label string `json:"label"`
	Hash  string `json:"sha1"`
	Pwned bool   `json:"pwned"`
	Count int    `json:"count,omitempty"`
	Err   string `json:"error,omitempty"`
}

func hashPassword(password string) string {
	h := sha1.Sum([]byte(password))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func queryHIBPPasswords(ctx context.Context, prefix string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hibpPasswordsAPI+prefix, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Add-Padding", "true")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func checkSuffix(apiResponse, suffix string) (bool, int) {
	scanner := bufio.NewScanner(strings.NewReader(apiResponse))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		if parts[0] == suffix {
			count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			return true, count
		}
	}
	return false, 0
}

func checkPassword(ctx context.Context, password string, idx int) PasswordResult {
	label := fmt.Sprintf("password #%d", idx)
	full := hashPassword(password)
	res := PasswordResult{Label: label, Hash: full}

	body, err := queryHIBPPasswords(ctx, full[:5])
	if err != nil {
		res.Err = err.Error()
		return res
	}

	found, count := checkSuffix(body, full[5:])
	res.Pwned = found
	res.Count = count
	return res
}

func checkPasswords(ctx context.Context, passwords []string) []PasswordResult {
	results := make([]PasswordResult, len(passwords))
	sem := make(chan struct{}, maxConcurrentPasswordReqs)
	var wg sync.WaitGroup

	for i, pw := range passwords {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, pw string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = checkPassword(ctx, pw, i+1)
		}(i, pw)
	}

	wg.Wait()
	return results
}
