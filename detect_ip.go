package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

func fetchIP(url string) (net.IP, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response (%s): %s", resp.Status, body)
	}
	ip := net.ParseIP(string(body))
	if ip == nil {
		return nil, fmt.Errorf("invalid response: '%s'", body)
	}
	return ip, nil
}

// detectIPv4 will try to resolve public facing IPv4 address, and return
// error if not possible
func detectIPv4() (net.IP, error) {
	return fetchIP("https://api.ipify.org")
}

// detectIPv64 will try to resolve public facing IPv6 address, and return
// error if not possible
func detectIPv64() (net.IP, error) {
	ip, err := fetchIP("https://api64.ipify.org")
	if err != nil {
		return nil, err
	}
	if ip.To4() != nil {
		return nil, fmt.Errorf("no IPv6 detected")
	}
	return ip, nil
}
