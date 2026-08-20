package main

import "net/http"
import "time"

const userAgent = "breach-checker/" + version + " (+https://github.com/p4p2r0/breach-checker)"

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}
