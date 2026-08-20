package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	hibpBreachAPI     = "https://haveibeenpwned.com/api/v3/breachedaccount/"
	emailRequestDelay = 1600 * time.Millisecond
)

type BreachInfo struct {
	Name        string   `json:"Name"`
	Title       string   `json:"Title"`
	Domain      string   `json:"Domain,omitempty"`
	BreachDate  string   `json:"BreachDate,omitempty"`
	PwnCount    int      `json:"PwnCount,omitempty"`
	DataClasses []string `json:"DataClasses,omitempty"`
}

type EmailResult struct {
	Email    string       `json:"email"`
	Breaches []BreachInfo `json:"breaches,omitempty"`
	Count    int          `json:"breachCount"`
	Err      string       `json:"error,omitempty"`
}

func checkEmail(ctx context.Context, email, apiKey string) EmailResult {
	res := EmailResult{Email: email}

	if apiKey == "" {
		res.Err = "no HIBP API key configured (--api-key or HIBP_API_KEY); get one at https://haveibeenpwned.com/API/Key"
		return res
	}

	body, status, err := fetchBreaches(ctx, email, apiKey)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	switch status {
	case http.StatusOK:
		var breaches []BreachInfo
		if err := json.Unmarshal(body, &breaches); err != nil {
			res.Err = fmt.Sprintf("could not parse API response: %v", err)
			return res
		}
		res.Breaches = breaches
		res.Count = len(breaches)
	case http.StatusNotFound:
	case http.StatusUnauthorized:
		res.Err = "unauthorized: invalid or missing HIBP API key"
	case http.StatusBadRequest:
		res.Err = "bad request: invalid email address"
	case http.StatusTooManyRequests:
		res.Err = "rate limited by HIBP API; try again shortly"
	default:
		res.Err = fmt.Sprintf("API returned status %d", status)
	}
	return res
}

func fetchBreaches(ctx context.Context, email, apiKey string) ([]byte, int, error) {
	endpoint := hibpBreachAPI + url.PathEscape(email) + "?truncateResponse=false"

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("hibp-api-key", apiKey)
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, 0, err
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, 0, readErr
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			wait := 2 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		return body, resp.StatusCode, nil
	}

	return nil, http.StatusTooManyRequests, fmt.Errorf("rate limited by HIBP API; try again shortly")
}

func checkEmails(ctx context.Context, emails []string, apiKey string) []EmailResult {
	results := make([]EmailResult, len(emails))
	for i, email := range emails {
		results[i] = checkEmail(ctx, email, apiKey)
		if i != len(emails)-1 {
			select {
			case <-ctx.Done():
				return results
			case <-time.After(emailRequestDelay):
			}
		}
	}
	return results
}
